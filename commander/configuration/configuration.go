package configuration

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/mushkevych/9ofm/commander/model"
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
		Add("log.enabled", "false").

		Add("diff.hide", "Modified,Added,Removed").

		Add("filetree.collapse-dir", "false").
		Add("filetree.pane-width", "0.5").
		Add("filetree.show-attributes", "true").
		Build()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// set global defaults (for performance)
	model.GlobalFileTreeCollapse = Config.GetBool("filetree.collapse-dir")
}
