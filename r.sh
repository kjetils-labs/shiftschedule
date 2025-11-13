envfile=".env"

if [ ! -f "$envfile" ]; then
  cp -v .env.example "$envfile";
fi

if ! test -f "$envfile"; then
  echo "env file $envfile does not exist, exiting ..."
  exit 1
fi

if docker compose > /dev/null 2>&1; then
  compose_cmd="docker compose"
elif command -v docker-compose; then
  compose_cmd="docker-compose"
else
  echo "Error: Docker Compose is not installed. Please install it and try again."
  exit 1
fi
version=$($compose_cmd version --short | cut -d "." -f 1)
if [ "$version" -lt 2 ]; then
    echo "Docker Compose version is to low, needs to be v2.0.0 or higher."
    exit 1
fi

$compose_cmd --env-file $envfile up postgres
