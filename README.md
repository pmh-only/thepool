# thepool.
thepool is a file sharing service with unlimited file sizes and reliable parallel upload/download capabilities.

## Try it now
![](/docs/screenshot.png)

:link: https://pool.pmh.codes

Since this service is exclusive to members, you can request an invite from an existing member.

## Concepts
- There are no restrictions on the size of files you can upload or download.
- Files are split into smaller chunks for upload/download, allowing for parallel transfers and resumption of interrupted transfers.
- If total size of chunk pool exceeds available storage space, least recently written chunks are deleted first.
- User can configure parallelism and chunk size to optimize transfer speeds based on their network conditions.
- Utilizes HTTP/2 for efficient multiplexing of requests.

## Installation
### Prerequisites
- MySQL 8.0+
- Docker or any other container runtime
- Publicly accessible minio endpoint for serving chunks

### Database Setup
1. Migrate your database by using [SQL file](/database.sql)
2. Make sure to create a user with appropriate permissions for the database.

### Minio Setup
1. Create a bucket in your minio instance to store file chunks.
2. Make sure the bucket is publicly accessible for reading chunks. Do not allow public write access.
3. Generate access key and secret key. Allow read and write permissions for the bucket.
4. You need private and public endpoint URLs for the minio instance.
    - Private endpoint is used by thepool server to read/write chunks.
    - Public endpoint is used by browsers to download chunks.
    - Here are some for example:
      - Private endpoint: `http://minio.internal:9000`
      - Public endpoint: `https://minio.example.com`

### Configure environment variables
1. Copy the example [environment file](/.env.example) to `.env`
2. Update the database connection settings in the `.env` file to match your database/minio configuration.

### Generate TLS certificates
This is required since thepool only works over HTTP/2 which requires HTTPS.

You can use openssl to generate self-signed certificates for testing purposes:
```
openssl req -x509 -newkey rsa:2048 \
  -keyout key.pem -out cert.pem \
  -days 36500 -nodes -subj "/CN=example.com"
```

Common Name (CN) does not have to match your domain name since it is only for testing purposes. \
For production, it is recommended to use certificates from a trusted CA like Let's Encrypt.

### Run thepool
You can run thepool using Docker:

```
docker run 
  -dp 8080:8080 --env-file .env \
  -v ./key.pem:/tmp/key.pem \
  -v ./cert.pem:/tmp/cert.pem \
  ghcr.io/pmh-only/thepool:latest
```

Server will be accessible at `https://localhost:8080`

## License
This project is licensed under the MIT License. See the [LICENSE](/LICENSE) file for details.
