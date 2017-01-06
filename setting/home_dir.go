package setting

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const XdgConfigHomeEnvKey = "XDG_CONFIG_HOME"
const HomeDirName = "popuko"

func HomeDir(base string) (bool, string) {
	var dir string
	if base == "" {
		dir = getXdgConfigHome() + "/" + HomeDirName
	} else {
		dir = base
	}

	root, err := filepath.Abs(dir)
	if err != nil {
		log.Println("error: cannot get the path to config home dir.")
		return false, ""
	}

	log.Println("info: Use " + root + " as the config root dir of this application.")
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
