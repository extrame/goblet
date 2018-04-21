package npm

import (
	"os"
	"os/exec"

	"encoding/json"

	"io/ioutil"

	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/goblet"
)

type NpmConfig struct {
	Name         string            `json:"name"`
	Dependencies map[string]string `json:"dependencies"`
}

type Npm struct {
	conf *NpmConfig
}

func (s *Npm) Init(server *goblet.Server) (err error) {
	server.AddFunc("npm", func(pkg, version string) error {
		if server.Env() == goblet.DevelopEnv {
			s.config(pkg, version)
			return s.runFor(pkg, version)
		}
		return nil
	})
	if s.conf == nil {
		s.conf = new(NpmConfig)
		s.conf.Dependencies = make(map[string]string)
	}
	//npm install -g browserify
	// if server.Env() == goblet.DevelopEnv {
	// 	exec.Command("npm", "install", "-g", "browserify").Run()
	// }
	s.conf.Name = server.Name
	return nil
}

func (n *Npm) config(pkg, version string) {
	if _, ok := n.conf.Dependencies[pkg]; !ok {
		n.conf.Dependencies[pkg] = version
	}
	go n.saveConfig()
}

func (n *Npm) saveConfig() {
	if bts, err := json.Marshal(n.conf); err == nil {
		ioutil.WriteFile("package.json", bts, 0666)
	}
}

func (n *Npm) runFor(pkg, version string) error {
	return exec.Command("npm", version).Run()
}

func (n *Npm) ParseConfig(prefix string) (err error) {
	dir := toml.String(prefix+".dir", "")
	if *dir != "" {
		if err = os.Chdir(*dir); err != nil {
			return err
		}
	}
	var bts []byte
	if bts, err = ioutil.ReadFile("package.json"); err == nil {
		n.conf = new(NpmConfig)
		err = json.Unmarshal(bts, n.conf)
	}
	return err
}
