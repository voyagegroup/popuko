package setting

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const XdgConfigHomeEnvKey = "XDG_CONFIG_HOME"
const homeDirName = "popuko"

func HomeDir(base string) (bool, string) {
	if base == "" {
		log.Println("info: Use $" + XdgConfigHomeEnvKey + " as the config root dir of this application.")
		base = getXdgConfigHome()
	}

	root, err := filepath.Abs(base + "/" + homeDirName)
	if err != nil {
		log.Println("error: cannot get the path to config home dir.")
		return false, ""
	}

	return true, root
}

func getXdgConfigHome() string {
	v := os.Getenv(XdgConfigHomeEnvKey)
	if v == "" {
		log.Println("info: try to use `~/.config` as $XDG_CONFIG_HOME")

		home, err := getHome()
		if err != nil {
			log.Fatal(err)
		}

		l, err := filepath.Abs(home + "/.config")
		if err != nil {
			log.Fatal(err)
		}

		v = l
	}

	path, err := filepath.Abs(v)
	if err != nil {
		log.Fatal(err)
	}

	return path
}

func getHome() (string, error) {
	isWin := runtime.GOOS == "windows"
	var HomeKey string
	if isWin {
		HomeKey = "USERPROFILE"
	} else {
		HomeKey = "HOME"
	}

	h := os.Getenv(HomeKey)
	home, err := filepath.Abs(h)
	if home == "" {
		err = errors.New("not found $" + HomeKey)
	}

	return home, err
}
