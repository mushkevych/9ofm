package system

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/ufoscout/go-up"
	"os"
	"path"
)

var Config go_up.GoUp

func init() {
	var err error

	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// $HOME/.config/.9ofm.yaml
	configFilePath := path.Join(home, ".config", ".9ofm.yaml")

	ignoreFileNotFound := true
	Config, err = go_up.NewGoUp().
		AddFile(configFilePath, ignoreFileNotFound).
		AddReader(go_up.NewEnvReader("", false, false)). // Loading environment variables
		Add("log.level", log.InfoLevel.String()).
		Add("log.path", "./9ofm.log").
		Add("debug", "false").
		Add("log.enabled", "true").

		Add("diff.hide", "Modified,Added,Removed").
		Build()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
