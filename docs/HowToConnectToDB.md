## How-To Connect To Our DB

We are using DocumentDB. This is a No-SQL DB based on MongoDB provided by AWS and is being hosted on
our DEV, QA and PROD cluster for the corresponding roles.

### Connection

#### CLI

For connection and working with the DB on your Terminal firstly install the [Mongo Client](https://docs.mongodb.com/manual/administration/install-community/)
on your host system and set all the necessary environment variables.

Afterwards, get the RDS combined ca bundle PEM file. You can use a simple cURL command for that.

`curl -X GET https://s3.amazonaws.com/rds-downloads/rds-combined-ca-bundle.pem -o ~/rds-combined-ca-bundle.pem`

Now, you have met all the pre-conditions for connecting to our glorious DB.

To do so, open your Terminal and enter the following command:

`mongo --ssl --host ae-mfm-yep.cluster-$CLUSTER_ID.eu-central-1.docdb.amazonaws.com:27017 --sslCAFile ~/rds-combined-ca-bundle.pem --username cluster_admin --password $CLUSTER_PASSWORD`

> Note:
> Replace the $CLUSTER_ID and $CLUSTER_PASSWORD by its corresponding values for the environments DEV, QA and PROD
> Currently, you can find those values in [Vaultier](https://vaultier.prod.cloudhh.de/#/workspaces/w/mfm-2/vaults/v/web-integration-platform/cards/c/documentdb/secrets)

Useful mongo commands can be found [here](https://docs.mongodb.com/manual/reference/mongo-shell/).

After successful connection, stand up and dance. yep yep

#### GUI

To connect to our DB one could use the JetBrains integrated DB connection settings.

Another great tool for connection to the DB would be [Robo 3T](https://robomongo.org/download)

But feel free to choose the tool of your desire. Some alternatives can be found [here](https://alternativeto.net/software/robo-3t/).
