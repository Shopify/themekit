package cmd

// import (
//	"os"
//	"path/filepath"
//	"testing"

//	"github.com/stretchr/testify/assert"

//	"github.com/Shopify/themekit/cmd/ystore"
//	"github.com/Shopify/themekit/kit"
// )

// func TestNewFileRegistry(t *testing.T) {
//	config := newTestConfig()
//	store, err := newFileRegistry(config)
//	assert.Nil(t, err)
//	s, _ := ystore.New(filepath.Join(config.Directory, storeName))
//	assert.Equal(t, store.store, s)
// }

// func TestSet(t *testing.T) {
//	config := newTestConfig()
//	storePath := filepath.Join(config.Directory, storeName)
//	registry, _ := newFileRegistry(config)
//	defer registry.store.Drop()
//	_, err := os.Stat(storePath)
//	assert.NotNil(t, err)
//	assert.NotNil(t, registry.Set(kit.Asset{Key: "test"}))
//	assert.Nil(t, registry.Set(kit.Asset{Key: "test", UpdatedAt: "test"}))
//	_, err = os.Stat(storePath)
//	assert.Nil(t, err)
// }

// func TestGet(t *testing.T) {
//	registry, _ := newFileRegistry(newTestConfig())
//	defer registry.store.Drop()
//	_, err := registry.Get("test.txt")
//	assert.Nil(t, err)
//	_, err = registry.Get("")
//	assert.NotNil(t, err)
//	assert.Nil(t, registry.Set(kit.Asset{Key: "test.txt", UpdatedAt: "123456"}))
//	version, err := registry.Get("test.txt")
//	assert.Nil(t, err)
//	assert.Equal(t, version, "123456")
// }

// func newTestConfig() *kit.Configuration {
//	config, _ := kit.NewConfiguration()
//	config.Environment = "test"
//	config.Domain = "test.myshopify.com"
//	config.ThemeID = "123"
//	config.Password = "sharknado"
//	return config
// }

// func TestExpandWildcards(t *testing.T) {
//	requestCount := make(chan int, 100)
//	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprintf(w, jsonFixture("responses/assets"))
//		requestCount <- 1
//	})
//	defer server.Close()

//	filenames, err := expandWildcards(client, []string{"assets/hello.txt"})
//	assert.Nil(t, err)
//	assert.Equal(t, len(requestCount), 0)
//	assert.Equal(t, filenames, []string{"assets/hello.txt"})

//	filenames, err = expandWildcards(client, []string{"assets/*"})
//	assert.Nil(t, err)
//	assert.Equal(t, len(requestCount), 1)
//	assert.Equal(t, filenames, []string{"assets/goodbye.txt", "assets/hello.txt"})
// }
