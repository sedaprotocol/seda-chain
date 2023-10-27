# This Dockerfile is for .gorelease purposesBuild: docker build -t seda-chaind .
# If you are looking for a Dockerfile to run a node, see dockerfiles/Dockerfile.node
FROM scratch
ENTRYPOINT ["/seda-chaind"]
COPY seda-chaind /
