#!/usr/bin/env bash
wget -O - http://download1.graphhopper.com/public/photon-db-latest.tar.bz2 |
bzip2 -cd | tar x
java -jar photon-0.2.5.jar -nominatim-import -host localhost -port 5432 -database nominatim -user postgres -password 123 -languages ru

java -jar photon-0.2.5.jar -host localhost -port 5432 -database nominatim -user postgres -password 123 -languages ru
