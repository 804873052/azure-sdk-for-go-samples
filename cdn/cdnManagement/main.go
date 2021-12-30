package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cdn/armcdn"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group2"
	profileName       = "sample2cdn2profile"
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
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	cdnProfile, err := createProfile(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("cdn profile:", *cdnProfile.ID)

	ssoUri, err := generateSsoURI(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Sso URI:", *ssoUri.SsoURIValue)

	// clean resource
	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createProfile(ctx context.Context, cred azcore.TokenCredential) (*armcdn.Profile, error) {
	cdnProfileClient := armcdn.NewProfilesClient(subscriptionID, cred, nil)
	pollerResp, err := cdnProfileClient.BeginCreate(
		ctx,
		resourceGroupName,
		profileName,
		armcdn.Profile{
			TrackedResource: armcdn.TrackedResource{
				Location: to.StringPtr("Global"),
			},
			SKU: &armcdn.SKU{
				Name: armcdn.SKUNamePremiumVerizon.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Profile, nil
}

func generateSsoURI(ctx context.Context, cred azcore.TokenCredential) (*armcdn.SsoURI, error) {
	cdnProfileClient := armcdn.NewProfilesClient(subscriptionID, cred, nil)
	resp, err := cdnProfileClient.GenerateSsoURI(ctx, resourceGroupName, profileName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SsoURI, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

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
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

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
