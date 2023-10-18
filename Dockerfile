# This Dockerfile si for .gorelease purposesBuild: docker build -t seda-chaind .
# If you are looking for a Dockerfile to run a node, see dockers/seda-node.Dockerfile
FROM scratch
ENTRYPOINT ["/seda-chaind"]
COPY seda-chaind /
