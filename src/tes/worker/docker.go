package tes_taskengine_worker

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

func NewDockerEngine() *DockerCmd {
	return &DockerCmd{}
}

type DockerCmd struct {
}

func (self DockerCmd) Run(containerName string, args []string,
	binds []string, workdir string, remove bool, jobId string,
  script *os.File, stdout *os.File, stderr *os.File) (int, error) {

	log.Printf("Docker Binds: %s", binds)

	docker_args := []string{"docker", "run", "--rm", "-i"}

	if workdir != "" {
		docker_args = append(docker_args, "-w", workdir)
	}

	for _, i := range binds {
		docker_args = append(docker_args, "-v", i)
	}

	// script mount
	local_script_path := script.Name()
	script_base := path.Base(local_script_path)
	mnt_script_path := path.Join("/tmp/", jobId, script_base)
	
	docker_args = append(docker_args, fmt.Sprintf("-v %s:%s:ro", local_script_path, mnt_script_path))

	// Write command to script file
	script.Write([]byte("#!/bin/bash\n"))
	script.Write([]byte(strings.Join(args, " ")))
	script.Close()	

  log.Printf("Command: %s", strings.Join(args, " "))  
	
	// docker args
  docker_args = append(docker_args, containerName)
	docker_args = append(docker_args, "/bin/bash")
	docker_args = append(docker_args, mnt_script_path)
	
	log.Printf("Runner: %s", strings.Join(docker_args, " "))

	cmd := exec.Command("/bin/bash", "-c", strings.Join(docker_args, " "))

	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}

	cmd_err := cmd.Run()
	exitStatus := 0
	if exiterr, ok := cmd_err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			exitStatus = status.ExitStatus()
			log.Printf("Exit Status: %d", exitStatus)	
			log.Printf("Error: %s", read_file_head(stderr.Name()))
		}
	} else {
		log.Printf("Exited Status: %d", exitStatus)
		//log.Printf("cmd.Run: %v", cmd_err)
	}

	return exitStatus, nil
}
