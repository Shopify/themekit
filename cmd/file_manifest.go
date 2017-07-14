package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/themekit/cmd/ystore"
	"golang.org/x/sync/errgroup"

	"github.com/Shopify/themekit/kit"
)

type fileManifest struct {
	store  *ystore.YStore
	mutex  sync.Mutex
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
	if manifest.local, err = store.Dump(); err != nil {
		return nil, err
	}

	if err := manifest.generateRemote(clients); err != nil {
		return nil, err
	}

	return manifest, manifest.prune(clients)
}

func (manifest *fileManifest) generateRemote(clients []kit.ThemeClient) error {
	manifest.remote = make(map[string]map[string]string)

	var requestGroup errgroup.Group
	for _, client := range clients {
		client := client
		requestGroup.Go(func() error {
			assets, err := client.AssetList()
			if err != nil {
				return err
			}
			for _, asset := range assets {
				manifest.mutex.Lock()
				if _, ok := manifest.remote[asset.Key]; !ok {
					manifest.remote[asset.Key] = make(map[string]string)
				}
				manifest.remote[asset.Key][client.Config.Environment] = asset.UpdatedAt
				manifest.mutex.Unlock()
			}
			return nil
		})
	}

	return requestGroup.Wait()
}

func (manifest *fileManifest) prune(clients []kit.ThemeClient) error {
	for filename := range manifest.local {
		if _, found := manifest.remote[filename]; !found {
			for _, client := range clients {
				path := filepath.ToSlash(filepath.Join(client.Config.Directory, filename))
				if info, err := os.Stat(path); err == nil && !info.IsDir() {
					found = true
					break
				}
			}
			if !found {
				if err := manifest.store.DeleteCollection(filename); err != nil {
					return err
				}
			}
		}
	}

	var err error
	manifest.local, err = manifest.store.Dump()
	return err
}

func parseTime(t string) time.Time {
	var parsed time.Time
	parsed, _ = time.Parse(time.RFC3339, t)
	return parsed
}

func (manifest *fileManifest) diffDates(filename, dstEnv, srcEnv string) (local, remote time.Time) {
	manifest.mutex.Lock()
	defer manifest.mutex.Unlock()

	if _, ok := manifest.local[filename]; ok {
		local = parseTime(manifest.local[filename][srcEnv])
	}
	if _, ok := manifest.remote[filename]; ok {
		remote = parseTime(manifest.remote[filename][dstEnv])
	}
	return local, remote
}

func (manifest *fileManifest) NeedsDownloading(filename, environment string) bool {
	localTime, remoteTime := manifest.diffDates(filename, environment, environment)
	return localTime.Before(remoteTime) || localTime.IsZero()
}

func (manifest *fileManifest) ShouldUpload(filename, environment string) bool {
	localTime, remoteTime := manifest.diffDates(filename, environment, environment)
	return remoteTime.Before(localTime) || remoteTime.IsZero()
}

func (manifest *fileManifest) ShouldRemove(filename, environment string) bool {
	localTime, remoteTime := manifest.diffDates(filename, environment, environment)
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
			} else {
				fetchableFilenames = append(fetchableFilenames, filename)
			}
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

func fmtTime(t time.Time) string {
	return "[" + t.Format("Jan 2 3:04PM 2006") + "]"
}

func (manifest *fileManifest) Diff(actions map[string]assetAction, dstEnv, srcEnv string) *themeDiff {
	diff := newDiff()
	for filename := range actions {
		local, remote := manifest.diffDates(filename, dstEnv, srcEnv)
		if !local.IsZero() && remote.IsZero() {
			diff.Removed = append(diff.Removed, kit.RedText(filename+" "+fmtTime(local)))
		}
		if local.IsZero() && !remote.IsZero() {
			diff.Created = append(diff.Created, kit.GreenText(filename+" "+fmtTime(remote)))
		}
		if !local.IsZero() && local.Before(remote) {
			diff.Updated = append(diff.Updated, kit.YellowText(filename+" local:"+fmtTime(local)+" remote:"+fmtTime(remote)))
		}
	}
	return diff
}

func (manifest *fileManifest) Set(filename, environment, value string) error {
	var err error
	manifest.mutex.Lock()
	defer manifest.mutex.Unlock()

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

	manifest.mutex.Lock()
	defer manifest.mutex.Unlock()
	if _, ok := manifest.remote[filename]; ok {
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
