package goblet

import (
	"github.com/sourcegraph/go-bower/bower"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var bower_cache = make(map[string]template.HTML)

func (s *Server) Bower(name string, version ...string) (h template.HTML, err error) {
	if *s.env == "production" {
		if res, ok := bower_cache[name]; ok {
			return res, nil
		}
	}

	root := filepath.Join(*s.wwwRoot, "public", "plugins", name)
	if _, err = os.Stat(root); os.IsNotExist(err) {
		if *s.env == "production" {
			log.Panicf("no %s plugins in production environment", name)
		}
		if _, err = os.Stat(filepath.Join(*s.wwwRoot, "public", ".bowerrc")); os.IsNotExist(err) {
			ioutil.WriteFile(filepath.Join(*s.wwwRoot, "public", ".bowerrc"), []byte(`{"directory" : "plugins"}`), 0644)
		}
		if len(version) > 0 {
			name = name + "#" + version[0]
		}
		c := exec.Command("bower", "install", "-S", name)
		c.Env = os.Environ()
		c.Dir = filepath.Join(*s.wwwRoot, "public")
		c.Stderr = LogFile
		if err = c.Run(); err != nil {
			return
		}
	}

	if bts, e := ioutil.ReadFile(filepath.Join(root, "bower.json")); e == nil {
		b, _ := bower.ParseBowerJSON(bts)
		res := appendHtml(s, b)
		h = template.HTML(res)
	} else {
		err = e
	}

	if *s.env == "production" {
		if err == nil {
			bower_cache[name] = h
		}
	}

	return
}

func appendHtml(s *Server, b *bower.Component) string {
	res := ""
	root := filepath.Join(*s.wwwRoot, "public", "plugins")
	log.Println(b)
	for k, _ := range b.Dependencies {
		if bts, e := ioutil.ReadFile(filepath.Join(root, k, "bower.json")); e == nil {
			if b1, err := bower.ParseBowerJSON(bts); err == nil {
				res += appendHtml(s, b1)
			} else {
				log.Println(err)
			}
		}
	}
	switch bs := b.Main.(type) {
	case []interface{}:
		for _, v := range bs {
			res += appendHtmlItem(*s.env, root, b.Name, v.(string))
		}
	case string:
		res += appendHtmlItem(*s.env, root, b.Name, bs)
	default:
		log.Panicf("%v,%T", b.Main, b.Main)
	}

	return res
}

func appendHtmlItem(env, root, name, v string) string {
	if strings.HasSuffix(v, ".js") {
		if env == "production" {
			//try to use min version
			v = strings.Replace(v, ".js", ".min.js", 1)
			if _, err := os.Stat(filepath.Join(root, name, v)); !os.IsNotExist(err) {
				return "<script src=/plugins/" + name + "/" + v + "></script>"
			}
		}
		return "<script src=/plugins/" + name + "/" + v + "></script>"
	} else if strings.HasSuffix(v, ".css") {
		return "<link href=/plugins/" + name + "/" + v + " rel='stylesheet'></link>"
	} else {
		return "<link href=/plugins/" + name + "/" + v + "></link>"
	}
}
