package cf

import(
	"cloud.google.com/go/functions/apiv2/functionspb"
	"context"
	functions "cloud.google.com/go/functions/apiv2"
	"fmt"
)

func GetCloudFunction(ctx context.Context, client *functions.FunctionClient, functionpath string) (*functionspb.GenerateDownloadUrlResponse, error) {

	req := &functionspb.GetFunctionRequest{
		Name: functionpath,
	}

	downloadReq := &functionspb.GenerateDownloadUrlRequest{
		Name: req.Name,
	}

	
	info, err := client.GetFunction(ctx, req)
	if err != nil {
		return nil, err
	}

	fmt.Println("Cloud Function details:", info)

	cloudFunction, err := client.GenerateDownloadUrl(ctx, downloadReq)
	if err != nil {
		return nil, err
	}
	return cloudFunction, nil

}