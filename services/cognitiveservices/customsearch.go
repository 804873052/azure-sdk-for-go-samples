// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package cognitiveservices

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/customsearch"
	"github.com/Azure/go-autorest/autorest"
)

func getCustomSearchClient(accountName string) customsearch.CustomInstanceClient {
	apiKey := getFirstKey(accountName)
	customSearchClient := customsearch.NewCustomInstanceClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	customSearchClient.Authorizer = csAuthorizer
	_ = customSearchClient.AddToUserAgent(config.UserAgent())
	return customSearchClient
}

//CustomSearch returns answers based on a custom search instance
func CustomSearch(accountName string) (*customsearch.WebWebAnswer, error) {
	customSearchClient := getCustomSearchClient(accountName)
	query := "Xbox"
	customConfig := "" // subsitute with custom config id configured at https://www.customsearch.ai

	searchResponse, err := customSearchClient.Search(
		context.Background(), // context
		customConfig,         // custom config (see comment above)
		query,                // query keyword
		"",                   // Accept-Language header
		"",                   // User-Agent header
		"",                   // X-MSEdge-ClientID header
		"",                   // X-MSEdge-ClientIP header
		"",                   // X-Search-Location header
		"",                   // country code
		nil,                  // count
		"",                   // market
		nil,                  // offset
		customsearch.Strict,  // safe search
		"",                   // set lang
		nil,                  // text decorations
		customsearch.Raw,     // text format
	)
	if err != nil {
		return nil, err
	}

	return searchResponse.WebPages, nil
}
