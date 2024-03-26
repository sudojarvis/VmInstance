## Dependencies

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

Credentials: Service Account credential key for authenticating with cloud compute and cloud functions.
Location: The location of the cloud function.
functionName: The name of the cloud function.
user: The user is username of the compute instance.
zone: The zone of the compute instance.

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
