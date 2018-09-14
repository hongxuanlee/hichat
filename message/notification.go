package message

import (
	"log"
	"os/exec"
	"runtime"
)

/**
 * when message come, send notification
 */

func Notify(name string, msg string) error {
	log.Print(runtime.GOOS)
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("osascript", "-e", "display notification \""+msg+"!\" with title \""+name+"!\"")
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
