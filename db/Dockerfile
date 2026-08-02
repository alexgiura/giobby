FROM postgres:16-alpine

# Bake init scripts into the image so production does not need a bind-mount.
COPY init_scripts/ /docker-entrypoint-initdb.d/
