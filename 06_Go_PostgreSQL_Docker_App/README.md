# Go - Full Stack App

## Start the Docker container.
docker run --name my_postgres_container -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_USER=myuser -e POSTGRES_DB=my_database -p 5432:5432 -d postgres

## Access the PostgreSQL terminal.
docker exec -it my_postgres_container psql -U myuser -d my_database
