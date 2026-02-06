for i in $(ls -1 /docker-entrypoint-initdb.d/*.sql); do psql languasia languasia -f $i; done;
