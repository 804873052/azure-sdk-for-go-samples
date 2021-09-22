package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	deploymentName    = "sample-deployment"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	exist, err := checkExistDeployment(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("deployment is exist:", exist)

	template, err := readJson("testdata/template.json")
	if err != nil {
		log.Fatal(err)
	}
	params, err := readJson("testdata/parameters.json")
	if err != nil {
		log.Fatal(err)
	}
	deploymentExtended, err := createDeployment(ctx, conn, template, params)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("created deployment:", *deploymentExtended.ID)

	validateResult, err := validateDeployment(ctx, conn, template, params)
	if err != nil {
		log.Fatal(err)
	}
	data, _ := json.Marshal(validateResult)
	log.Println("validate deployment:", string(data))

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func readJson(path string) (map[string]interface{}, error) {
	templateFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	template := make(map[string]interface{})
	if err := json.Unmarshal(templateFile, &template); err != nil {
		return nil, err
	}

	return template, nil
}

func checkExistDeployment(ctx context.Context, conn *arm.Connection) (bool, error) {
	deploymentsClient := armresources.NewDeploymentsClient(conn, subscriptionID)

	boolResp, err := deploymentsClient.CheckExistence(ctx, resourceGroupName, deploymentName, nil)
	if err != nil {
		return false, err
	}

	return boolResp.Success, nil
}

func createDeployment(ctx context.Context, conn *arm.Connection, template, params map[string]interface{}) (*armresources.DeploymentExtended, error) {
	deploymentsClient := armresources.NewDeploymentsClient(conn, subscriptionID)

	deploymentPollerResp, err := deploymentsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		deploymentName,
		armresources.Deployment{
			Properties: &armresources.DeploymentProperties{
				Template:   template,
				Parameters: params,
				Mode:       armresources.DeploymentModeIncremental.ToPtr(),
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create deployment: %v", err)
	}

	resp, err := deploymentPollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("cannot get the create deployment future respone: %v", err)
	}

	return &resp.DeploymentExtended, nil
}

func validateDeployment(ctx context.Context, conn *arm.Connection, template, params map[string]interface{}) (*armresources.DeploymentValidateResult, error) {
	deploymentsClient := armresources.NewDeploymentsClient(conn, subscriptionID)

	pollerResp, err := deploymentsClient.BeginValidate(
		ctx,
		resourceGroupName,
		deploymentName,
		armresources.Deployment{
			Properties: &armresources.DeploymentProperties{
				Template:   template,
				Parameters: params,
				Mode:       armresources.DeploymentModeIncremental.ToPtr(),
			},
		},
		nil)

	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return &resp.DeploymentValidateResult, nil
}

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroupsCreateOrUpdateResult.ResourceGroup, nil
}

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}