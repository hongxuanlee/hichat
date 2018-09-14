package message

import (
	"log"
	"os/exec"
	"runtime"
)

/**
 * when message come, send notification
 */

func Notify(msg string) error {
	log.Print(runtime.GOOS)
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("osascript", "-e", "display notification \"Hello world!\" with title \""+msg+"!\"")
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
