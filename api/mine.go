package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"text/template"

	"github.com/gin-gonic/gin"
)

func MineStart(c *gin.Context) {

	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error})
	}

	wallet := Wallet{}
	if err := json.Unmarshal(data, &wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl := template.Must(template.New("jobTemplate").Parse(JobTemplate))
	var jobScript bytes.Buffer
	if err := tmpl.Execute(&jobScript, wallet); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	}

	jobScriptFile, err := os.Create("mining_job.sh")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	defer jobScriptFile.Close()
	jobScriptFile.WriteString(jobScript.String())

	cmd := exec.Command("sh", "-c", "sbatch mining_job.sh")
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Mining job started"})
		output, _ := cmd.Output()

		jobIDRegex := regexp.MustCompile(`\d+`)
		jobID := string(jobIDRegex.Find(output))

		os.Setenv("MINING_JOB_ID", jobID)
	}
}

func MineStop(c *gin.Context) {

	cmd := exec.Command("sh", "-c", "scancel", os.Getenv("MINING_JOB_ID"))
	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Mining job stopped"})
	}
}
