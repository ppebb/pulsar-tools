package main

import (
	"fmt"
	"os"
	"os/exec"
)

// TODO: I hate invoking a binary for this, but compiling with lib-xbmg doesn't
// work (go sucks), and I don't feel like rimplementing xbmg myself yet
func DecodeBMG(bmg []byte) ([]byte, error) {
	bmgHandle, err := os.OpenFile("temp/bmg.bmg", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	bmgHandle.Write(bmg)
	bmgHandle.Close()

	cmd := exec.Command("./wbmgt", "decode", "bmg.bmg", "--no-header", "--export")
	cmd.Dir = "temp"

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		stderrPipe.Close()
		return nil, err
	}

	stderr := []byte{}
	n, err := stderrPipe.Read(stderr)

	// Closes stderrPipe, so nothing extra is needed
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	if n != 0 {
		return nil, fmt.Errorf("Error while running wbmgt decode! %s", stderr)
	}

	parsedHandle, err := os.Open("temp/bmg.txt")
	if err != nil {
		return nil, err
	}

	ret := []byte{}
	_, err = parsedHandle.Read(ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
