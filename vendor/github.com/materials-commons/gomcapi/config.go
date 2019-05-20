package mcapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/materials-commons/config"
	"github.com/materials-commons/config/cfg"
	"github.com/materials-commons/config/handler"
	"github.com/materials-commons/config/loader"
)

func init() {
	h := setupConfigHandler()
	config.Init(h)
}

// setupConfigHandler creates the handler for the mc package. It sets up a
// multi handler. If the user has setup a config.json in their .materialscommons
// directory then it will add that to the handler list. Handlers are searched in
// the following order: env - (optional) config file - defaults.
//
// This means that configuration set in the environment will override all other
// settings, then it will check the config file (if one is setup), and finally
// it will use the defaults.
func setupConfigHandler() cfg.Handler {
	u, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Couldn't determine current user: %s", err))
	}

	// Set up the handlers. The order matters as it will search for
	// configuration entries first to last, stopping when it finds
	// one. This means that each entry overrides settings below it.
	handlers := []cfg.Handler{
		handler.Env(),
	}

	configFile := filepath.Join(u.HomeDir, ".materialscommons/config.json")
	if l := getUserConfigLoader(configFile); l != nil {
		handlers = append(handlers, handler.Loader(l))
	}

	defaultHandler := handler.Map()
	loadDefaults(defaultHandler)
	handlers = append(handlers, defaultHandler)
	return handler.Sync(handler.Multi(handlers...))
}

// getUserConfigLoader returns a json loader if the $HOME/.materialscommons/config.json
// file exists. It will panic if the file exists but cannot be read.
func getUserConfigLoader(configFile string) cfg.Loader {
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		contents, err := ioutil.ReadFile(configFile)
		if err != nil {
			panic(fmt.Sprintf("%s exists but can't be read: %s", configFile, err))
		}
		return loader.JSON(bytes.NewReader(contents))
	}
	return nil
}

// loadDefaults sets up the default values for the following configuration keys:
//     mcurl: https://materialscommons.org/api
//     mclogging: info
func loadDefaults(h cfg.Handler) {
	h.Set("mcurl", "https://materialscommons.org/api")
	h.Set("mclogging", "info")
}
