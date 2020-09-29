# gen

maybe this:
https://github.com/fullstorydev/grpcui/blob/master/internal/resources/webform/gen.go

https://github.com/fullstorydev/grpcurl

## Modular

Each Module can do their own generation, and the Dev can run using their example.

- You can call the mini generator as a developer at any level.

- Caching is another option but i think is a nest of problems to get right. Instead maybe we introduce caching just inside bs-lang
	- bs-lang can have a Tranlation Memory are they can it in the industry.
	- It can have a genji DB for it. Why ? because we can later use it for runtime translation too. 
		- Ex: Like when 2 people are chating but oyu want it in your language
	- We could bake it into sys-core now.
		- then the Design time and Runtime uses it. Its not aweful.

## Flutter and GO

Go Generate can be used for Flutter and Golang ?
A generate.go placed in the flutter project should do it.

The bs-gen golang then calls the bs-lang flutter, etc etc

That generate.go needs however to import the generate.go of any sub modules in order to force the recursive generation. From MainTemplate the imports naturally force sub modules to be generated.
- SO each dev does a cd ./../mod-x go:generate

## Modular

Each Generator is a mini thing. go:generate in each Module calls the "bs-gen" golang code that has links to the mini generators. This keeps things flexible.


## Outputs

If the output is go embedded, then we dont have to worry about moving the file around.

If the generator gens a Manifest as part of the embedding, the runtime can use it when it starts up. We need 2 manifest:
- A golang one. So sys-core can reflect on it.
- A JSON one. So the Dev can see what happened

bs-data is easy. Data and Migrations get embedded, and sys-core can then reflect and access them.

bs-lang flutter is the hardest, but not sure yet.

## Runtime

Sys-core needs to use the generated code of each Module.

- Maintemplate can reflect off the embedded code and  pass the needed module names to Sys-core. We can remove this by having a gen-json file being embedded in MainTemplate that Sys-core can always look for !

- Sys-core can then activate the Migrations based on the sys-core DB timestamp (stored in the DB)