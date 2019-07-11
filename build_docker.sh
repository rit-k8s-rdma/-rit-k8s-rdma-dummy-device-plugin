if [[ -z $VERSION ]]; then 
    echo "Error VERSION env var is not set. Failing to continue.";
    exit 1
fi

docker build -t rdma-dummy-dp:$VERSION .
docker save rdma-dummy-dp:$VERSION > rdma-dummy-dp.tar