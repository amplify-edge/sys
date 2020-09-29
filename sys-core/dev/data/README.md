# Data

We need to do DB migrations, Test data and Bootstrap data in a unified way. Everyone does but they dont :)


## v2 Design

The migrations are run against the DB, and NOT against the golang struct.
- This is because a V2 migration in SQL, will break if it relies on a V3 golang model struct.

Use a UTC datetime stamp ( for ordering )
- The tool will do it for you. Part of the shared sys-main cli or bs-data .
- We will use a timestamped folder to make it clean for migration.

Some tools can reflect on the DB and gen the differences for you.
- The genji tool now has a dump command
	- https://github.com/genjidb/genji/pull/200
	- https://github.com/genjidb/genji/issues/181
	- So by comparing the CURRENT DB to the PREVIOUS ( full DB schema saved in that folder ) we can see the difference.
	- The dumb command is scoped by Table which we need for Modular Archi.
	- The dump command is designed to dump the schema and data btw !
		- Because some data you might want as part of your migration too.
		- We might be able to get them to make the tool in genji allow just dumping the schema too. We can aks them 
		- I dont think we should be dependent on their genji shell ?
			- Its too limiting for us, and we want to use GRPC and CLI so this can be done remotely. Its a bad design they have in that you cant run the shell and still have the DB running actually.
			- Our tool can import theirs and we can then hide what we dont want.
- SO all the Dev needs to do in each migration phase is to do a dump and compare and save it into the Migration CURRENT folder.
- SO they can see the difference visually in their IDE. I think for now writing a diff generator it too much. But i think always creating a dump schema as part of the tooling we should do. Its easy and enforced the approach.
- For now you hand write the SQL migration, and later we can gen it maybe using some of the genji code
- This means that our DesignTime toolng is dependent on Genjo code. So i think we need a "shared" repo acually.
	- Then the CLI for Mod devs and Sys devs is using the real DB.
	- We can see what Module they are working in too.

Use constants per migration in golang
- Will make the migration easier to test 
- When making a new migration, just copy the constants from the previous migration



The runtime gets the embedded migrations and just runs them.
- runtme know the timstamps in the DB. ALL timestamp are ALWAYS converted to GMT anyway.
- the runtime then just runs the migrations. Easy

Migration Process
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

---

## V3 Crazy Design
We use badgerdb and genji.

Good basis i think.
https://github.com/amacneil/dbmate

Later we will add change feed to the DB.
Later we will need to upgrade many DB's in unison.

The DB is embedded in a normal golang project or if you want you can just run it on its own, and call it over grpc. Later will be needed when we split each mod into GRPC and DATA layers, which is natural for scaling and Perf later.

SO you can do DB migrations using the sys CLI.
So the DB URL gets all the GRPC security and other things for free.
The same applies if the dB is embedded. It’s just GRPC after all.

We store the version time stamp in all the dB instances and so we can upgrade many DB's in unison using GRPC streaming. 

We don’t have the migrations in plain files. They are golang files and so get compiled into the code.
The constants are per migration - they HAVE TO BE. So the migrations are strongly typed. Much more secure and if compile fails then maybe it’s your migration code. This is what you want imho.
Constants must be hand coded, and not via Reflection. Just take a copy from the previous migration and change it as needed. YOu can see the change over time by looking in each folder - yes use a Folder with a timestamp name for each migration, its ordered and cleared and contained.
- What are the constants ? 
- DB Table
- DB index, field and types


So yeah it’s a pretty different approach.

Also there is the CLI and web GUI because all that is generated at compile time easily off the GRPC and the migrations are viewable by just reflecting off the code at runtime because the migrations are in golang :)
So Ops Team don’t have go say a pray before running a migration.

DB can be put into auto migrate or checked migrate state. In checked when a migration (due to CD deployment) is needed the ops team gets a notification and url to the web GUI. Also this is a natural fit for audit logging into the DB of the migrations, so that in CLI and GUI you can see what happened. Its just an event being recorded "

The same applies to test and bootstrap data. It’s all golang.
It’s part of each migration too of course because schema and test data must match.

SO the migrations are the key.
So now the golang struct Model is per migration. Reminds me of GRPC ? Field that can be ignored if not needed.
SO why dont we make our DAO models using Protobufs ? Perfect match for Genji because each Table knows nothing about other tables. Its tight. 
We load tons of test Data in.

So now go tests are Unified with migrations.