# Data

We need to do DB migrations, Test data and Bootstrap data in a unified way. Everyone does but they dont :)


## v2 Design

The migrations are run against the DB, and NOT against the golang struct.
- This is because a V2 migration in SQL, will break if it relies on a V3 golang model struct.

Use a UTC datetime stamp ( for ordering )
- The tool will do it for you. Part of the shared sys-main cli or bs-data .
- We will use a timestamped folder to make it clean for migration.

- The genji tool now has a dump command
	- https://github.com/genjidb/genji/pull/200
	- https://github.com/genjidb/genji/issues/181
	- So by comparing the CURRENT DB to the PREVIOUS ( full DB schema saved in that folder ) we can see the 


## Migration Process that sys core does

- phase 1
	- use the Test DB ( its just a different folder )
	- back it up
	- run migration
	- run test data in
	- run go tests to exercise the golang and DB to ensure the migration has not messed up things.
	- same for proj-data
- phase 2
	- use the real DB
	- so exactly the same thing....

	
## DEV does this !!
Migration Directory structure

Each Module has this.

- migration delta
	- zero data
		- sql.dumb ( everything )
		- do.sql
			- e.g create 2 new tables
		- transform.go
			- e.g split the data from 1 table to 2.
		- test-data.go
			- add any new data to make it whole.
			- go test onto go struts
			- data.json
		- proj-data.go
			- add any new data to make it whole
			- go test onto go struts
			- data.json
	- 20201002-xxx
		- sql.dumb ( everything )
		- do.sql
			- e.g create 2 new tables
		- transform.go
			- e.g split the data from 1 table to 2.
		- test-data.go
			- add any new data to make it whole
		- proj-data.go
			- add any new data to make it whole
	- LATEST
		- sql-dumb

## WHERE and TODO

- sys-core needs to model and provide a TestDB
- sys-core needs a Table called "migration", to hold
	- UTC dateTime stamp
- sys-core needs a Table called "migration-audit", to hold
	- log of the migrations.

# sys-core cli
- Used at DESGN TIME and Runtime.
- cli over GRPC (sys-core)
	- extend the cli to do the local design time things also.
	- creates the folders for a module with the right DT
	- calls the genji dump ( with correct Tables )
		- call over grpc to get the dumb
		- put it into the right dev folder.


