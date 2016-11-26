# ableto-engineering-2017

Behaviorix was created for AbleTo's Summer 2017 Engineering Challenge. <br>
Deployed to Google Cloud at https://behaviorix.appspot.com.

Tested with Docker version 1.12.3. 
## Run locally

To run webapp locally, execute the following commands:

```
docker built -t my-app-name .
docker run -it —-rm -p 8000:8000 my-app-name
```

## Testing

To run tests and teardown, execute the following commands:

```
docker-compose -f docker-compose.test.yml up
docker-compose -f docker-compose.test.yml down
```
