package integrationtests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestMattermost(t *testing.T) {
	/* Test case for deploying a matermost.

	   **Test Scenario**

	   - Deploy a matermost.
	   - Check that the outputs not empty.
	   - Check that vm is reachable.
	   - Make sure mattermost zinit service is running.
	   - Destroy the deployment.
	*/

	// retryable errors in terraform testing.
	// generate ssh keys for test
	publicKey, privateKey, err := GenerateSSHKeyPair()
	if err != nil {
		t.Fatalf("failed to generate ssh key pair: %s", err.Error())
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./mattermost",
		Vars: map[string]interface{}{
			"public_key": publicKey,
		},
		Parallelism: 1,
	})
	defer terraform.Destroy(t, terraformOptions)

	_, err = terraform.InitAndApplyE(t, terraformOptions)
	assert.NoError(t, err)

	// Check that the outputs not empty
	yggIP := terraform.Output(t, terraformOptions, "ygg_ip")
	assert.NotEmpty(t, yggIP)
	fqdn := terraform.Output(t, terraformOptions, "fqdn")
	assert.NotEmpty(t, fqdn)

	ok := TestConnection(yggIP, "22")
	assert.True(t, ok)

	// Check that the solution is running successfully
	output, err := RemoteRun("root", yggIP, "zinit list", privateKey)
	assert.NoError(t, err)
	assert.Contains(t, output, "mattermost: Running")

	statusOk := false
	ticker := time.NewTicker(2 * time.Second)
	for now := time.Now(); time.Since(now) < 1*time.Minute; {
		<-ticker.C
		resp, err := http.Get(fmt.Sprintf("http://%s", fqdn))
		if err == nil && resp.StatusCode == http.StatusOK {
			statusOk = true
			break
		}
	}
	assert.True(t, statusOk, "website did not respond with 200 status code")
}