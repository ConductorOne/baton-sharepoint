FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-sharepoint"]
COPY baton-sharepoint /