// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	serverName        = "sampleyserver"
	configurationName = "sample-postgresql-configuration"
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

	server, err := createServer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql server:", *server.ID)

	configuration, err := createConfiguration(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgresql configuration:", *configuration.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context, cred azcore.TokenCredential) (*armpostgresql.Server, error) {
	serversClient, err := armpostgresql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := serversClient.BeginCreate(
		ctx,
		resourceGroupName,
		serverName,
		armpostgresql.ServerForCreate{
			Location: to.Ptr(location),
			Properties: &armpostgresql.ServerPropertiesForDefaultCreate{
				CreateMode:                 to.Ptr(armpostgresql.CreateModeDefault),
				InfrastructureEncryption:   to.Ptr(armpostgresql.InfrastructureEncryptionDisabled),
				PublicNetworkAccess:        to.Ptr(armpostgresql.PublicNetworkAccessEnumEnabled),
				Version:                    to.Ptr(armpostgresql.ServerVersionEleven),
				AdministratorLogin:         to.Ptr("dummylogin"),
				AdministratorLoginPassword: to.Ptr("QWE123!@#"),
			},
			SKU: &armpostgresql.SKU{
				Name: to.Ptr("B_Gen5_1"),
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
	return &resp.Server, nil
}

func createConfiguration(ctx context.Context, cred azcore.TokenCredential) (*armpostgresql.Configuration, error) {
	configurationsClient, err := armpostgresql.NewConfigurationsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := configurationsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		configurationName,
		armpostgresql.Configuration{
			Properties: &armpostgresql.ConfigurationProperties{
				Source: to.Ptr("user-override"),
				Value:  to.Ptr("off"),
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
	return &resp.Configuration, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}