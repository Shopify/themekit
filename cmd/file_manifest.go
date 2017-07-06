package cmd

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/themekit/cmd/internal/ystore"
	"golang.org/x/sync/errgroup"

	"github.com/Shopify/themekit/kit"
)

type fileManifest struct {
	store  *ystore.YStore
	local  map[string]map[string]string
	remote map[string]map[string]string
}

const storeName = "theme.lock"

func newFileManifest(path string, clients []kit.ThemeClient) (*fileManifest, error) {
	store, err := ystore.New(filepath.Join(path, storeName))
	if err != nil {
		return nil, err
	}

	manifest := &fileManifest{store: store}

	manifest.local, err = store.Dump()
	if err != nil {
		return nil, err
	}

	return manifest, manifest.generateRemote(clients)
}

func (manifest *fileManifest) generateRemote(clients []kit.ThemeClient) error {
	manifest.remote = make(map[string]map[string]string)
	var mutex sync.Mutex

	var requestGroup errgroup.Group
	for _, client := range clients {
		client := client
		requestGroup.Go(func() error {
			assets, err := client.AssetList()
			if err != nil {
				return err
			}
			mutex.Lock()
			for _, asset := range assets {
				if _, ok := manifest.remote[asset.Key]; !ok {
					manifest.remote[asset.Key] = make(map[string]string)
				}
				manifest.remote[asset.Key][client.Config.Environment] = asset.UpdatedAt
			}
			mutex.Unlock()
			return nil
		})
	}

	return requestGroup.Wait()
}

func parseTime(t string) time.Time {
	var parsed time.Time
	parsed, _ = time.Parse(time.RFC3339, t)
	return parsed
}

func (manifest *fileManifest) diffDates(filename, dstEnv, srcEnv string) (local, remote time.Time) {
	if _, ok := manifest.local[filename]; ok {
		local = parseTime(manifest.local[filename][srcEnv])
	}
	if _, ok := manifest.remote[filename]; ok {
		remote = parseTime(manifest.remote[filename][dstEnv])
	}
	return local, remote
}

func (manifest *fileManifest) fileDates(filename, env string) (local, remote time.Time) {
	return manifest.diffDates(filename, env, env)
}

func (manifest *fileManifest) NeedsDownloading(filename, environment string) bool {
	localTime, remoteTime := manifest.fileDates(filename, environment)
	return localTime.Before(remoteTime) || localTime.IsZero()
}

func (manifest *fileManifest) ShouldUpload(filename, environment string) bool {
	localTime, remoteTime := manifest.fileDates(filename, environment)
	return remoteTime.Before(localTime) || remoteTime.IsZero()
}

func (manifest *fileManifest) ShouldRemove(filename, environment string) bool {
	localTime, remoteTime := manifest.fileDates(filename, environment)
	return remoteTime.Before(localTime)
}

func (manifest *fileManifest) Should(event kit.EventType, filename, environment string) bool {
	if event == kit.Update || event == kit.Create {
		return manifest.ShouldUpload(filename, environment)
	} else if event == kit.Remove {
		return manifest.ShouldRemove(filename, environment)
	} else if event == kit.Retrieve {
		return manifest.NeedsDownloading(filename, environment)
	}
	return false
}

func (manifest *fileManifest) FetchableFiles(filenames []string, env string) []string {
	fetchableFilenames := []string{}
	if len(filenames) <= 0 {
		for assetName := range manifest.remote {
			fetchableFilenames = append(fetchableFilenames, assetName)
		}
	} else {
		wildCards := []string{}
		for _, filename := range filenames {
			if strings.Contains(filename, "*") {
				wildCards = append(wildCards, filename)
			}
			fetchableFilenames = append(fetchableFilenames, filename)
		}

		if len(wildCards) > 0 {
			for assetName := range manifest.remote {
				for _, wildcard := range wildCards {
					if matched, _ := filepath.Match(wildcard, assetName); matched {
						fetchableFilenames = append(fetchableFilenames, assetName)
					}
				}
			}
		}
	}
	return fetchableFilenames
}

func (manifest *fileManifest) Diff(actions map[string]assetAction, dstEnv, srcEnv string) *themeDiff {
	diff := newDiff()
	for filename := range actions {
		local, remote := manifest.diffDates(filename, dstEnv, srcEnv)
		if !local.IsZero() && remote.IsZero() {
			diff.Removed = append(diff.Removed, kit.RedText(filename))
		}
		if local.IsZero() && !remote.IsZero() {
			diff.Created = append(diff.Created, kit.GreenText(filename))
		}
		if !local.IsZero() && local.Before(remote) {
			diff.Updated = append(diff.Updated, kit.YellowText(filename))
		}
	}
	return diff
}

func (manifest *fileManifest) Set(filename, environment, value string) error {
	var err error
	if _, ok := manifest.remote[filename]; !ok {
		manifest.remote[filename] = make(map[string]string)
	}
	manifest.remote[filename][environment] = value

	batch := manifest.store.Batch()
	for env, version := range manifest.remote[filename] {
		currentVersion, _ := manifest.store.Read(filename, environment)
		if currentVersion == "" || env == environment {
			if err = batch.Write(filename, env, version); err != nil {
				return err
			}
		}
	}

	if err = batch.Commit(); err != nil {
		return err
	}

	manifest.local, err = manifest.store.Dump()
	return err
}

func (manifest *fileManifest) Delete(filename, environment string) error {
	if err := manifest.store.Delete(filename, environment); err != nil {
		return err
	}

	if _, ok := manifest.remote[filename]; !ok {
		delete(manifest.remote[filename], environment)
	}

	var err error
	manifest.local, err = manifest.store.Dump()
	return err
}

func (manifest *fileManifest) Get(filename, environment string) (string, error) {
	version, err := manifest.store.Read(filename, environment)
	if err != nil && err != ystore.ErrorCollectionNotFound && err != ystore.ErrorKeyNotFound {
		return "", err
	}
	return version, nil
}
