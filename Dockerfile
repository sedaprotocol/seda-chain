# This Dockerfile is for .gorelease purposesBuild: docker build -t sedad .
# If you are looking for a Dockerfile to run a node, see dockerfiles/Dockerfile.node
FROM scratch
ENTRYPOINT ["/sedad"]
COPY sedad /
