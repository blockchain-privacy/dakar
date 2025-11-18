# Dgraph Upgrade Guide

To upgrade the database to new version all data must be first exported with the current version and imported with the
new version.

## Export data

More information [here](https://dgraph.io/docs/deploy/dgraph-administration/#exporting-database).

1. Stop all clients accessing Dgraph (including Dakar).

1. With the current Dgraph version, export the database. Issue the following GraphQL request to the admin endpoint of the alpha
   node ``http://localhost:8080/admin``.

    ```graphql
    mutation {
      export(input: {format: "rdf"}) {
        response {
          message
          code
        }
      }
    }
    ```
1. Follow the export process in the docker log via ``docker logs -f alpha``.
1. After the export process is finished the new directory `export` contains the exported data.

## Import data via bulk loader

More information [here](https://dgraph.io/docs/deploy/fast-data-loading/bulk-loader/).

1. Get the new version of Dgraph
1. With a new directory start a zero node
1. Use ```docker/docker-compose-bulk.yml``` with the correct arguments
    1. Set the correct file path to the directory containing the exported data
    1. Set the correct filenames in the arguments of the bulk loader
1. Follow the progress of the bulk loader via `docker logs -f bulk`
1. A new directory `out` is created, with the following structure:
   ```text
   ./out
   ├── 0
   │   └── p
   │       ├── 000000.vlog
   │       ├── 000002.sst
   │       └── MANIFEST
   ```
1. Move the `p` directory to the new folder from Step 2.
1. Start Dgraph as usual
