# Policy

This is where the actual authorization of the resources will be provided.

- We could query the GenjiDB via SQL (Policy are stored inside the database), minimizing bottleneck between actual persistence and policy enforcement.
- We could embed something like OPA (Rego) or Casbin as policy enforcer, each module will then have to store their own policy inside their own table namespace.
- Once loaded either partially or wholly, any request to given module will then have to pass that specific enforcement policy.

## TODO

- [] Investigate embedding and enforcing OPA policy via SQL
- [] Investigate embedding and enforcing Casbin policy via available SQL / badger adapter.