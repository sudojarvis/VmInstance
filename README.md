## Dependencies

- gofiber: Web framework for Go
- googleapis/google-cloud-go: Google Cloud client libraries for Go
- google.golang.org/api/compute/v1: Google Compute Engine API client library for Go
- cloud.google.com/go/compute/apiv1: Google Compute Engine API client library for Go.
- google.golang.org/api/option: Google API client library for Go.
- google.golang.org/genproto/googleapis/cloud/compute/v1: Protobuf definitions for Google Cloud Compute Engine API.
- google.golang.org/protobuf/proto: Go support for Protocol Buffers.
- cloud.google.com/go/functions/apiv2/functionspb: Google Cloud Functions API client library for Go.
- cloud.google.com/go/functions/apiv2: Google Cloud Functions client library for Go.
- github.com/pkg/sftp: Go package for SFTP client.
- golang.org/x/crypto/ssh: Go package for SSH client.

## Usage

1. Set up Google Cloud Credentials

   - Create a service account key with the necessary permissions for accessing GCP resources.
   - Download the JSON key file and securely store it.
   - Run the following command to authenticate gcp with service account

   ```
     gcloud auth activate-service-account --key-file=path/to/service-account-key.json
   ```

2. Build and Run the Application

- Run the following command to build the application

```
  docker-compose up --build
```

- To go inside the container, run the following command

```
  sudo docker exec -it gcp_scanner /bin/bash
```

- To stop the application, run the following command

```
  docker-compose down
```

3. Access the application endpoints
   - The application listens on port 3000.
   - Send a POST request to <http://0.0.0.0:3000/test> with the required parameters in the request body.

## Request Body

This will contain credentials including details of Service Account, SSH Key and Cloud Function being scanned.

- Credentials: Service Account credential key for authenticating with cloud compute and cloud functions.
  
- Location: The location of the cloud function.
  
- functionName: The name of the cloud function.
  
- user: The user is username of the compute instance for ssh connection.
  
- zone: The zone of the compute instance.

Use the following format for request body:

      {
        "credentials": {
          "type": "service_account",
          "project_id": "",
          "private_key_id": "",
          "private_key": "",
          "client_email": "",
          "client_id": "",
          "auth_uri": "https://accounts.google.com/o/oauth2/auth",
          "token_uri": "https://oauth2.googleapis.com/token",
          "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
          "client_x509_cert_url": "",
          "universe_domain": "googleapis.com"
        },
        "Location": "",
        "functionName": "",
        "user": "",
        "zone": ""
      }

