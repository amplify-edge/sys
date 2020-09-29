# Config


Each Mod as their own config

config.yaml

Boostrap
Use config.mk in shared/boilerplate


At Design Time:
- in you server, Put your config model with the config.yaml
- in your main, 
	- use a main.go and wire it and the sub module up.
	- modify your makefile (this-all), to do the requires make calls.
		- this-build-all; does the same this to the sub module
		- See the repo main/deploy/maintemlate


At GenTime:
- tell you and your sub module to
- copy them al up to you main
- merge them into a base.yaml
- and override what you want


