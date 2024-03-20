package main

import(
	"fmt"
	// "log"
	// "os"
	"os/exec"
)

func checkSemgrepVersion() error {
	cmd := exec.Command("semgrep", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error checking Semgrep version: %v", err)
	}

	fmt.Printf("Semgrep version:\n%s\n", output)
	return nil
}

func installSemgrep() error{
	// Install Semgrep using pip
	cmd := exec.Command("python3","-m","pip", "install", "semgrep")
	if err := cmd.Run(); err != nil {
		return err
	}

	// remove the installation script after use
	// if err := os.Remove("install_semgrep.sh"); err != nil {
	// 	return err
	// }

	return nil
}

func runSemgrep() error{
	
	cmd := exec.Command("semgrep", "scan", "--json", "--output=report.json")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func loginSemgrep() error{
	cmd := exec.Command("semgrep", "login")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}