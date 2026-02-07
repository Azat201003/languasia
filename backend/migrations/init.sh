for i in $(ls -1 ./*.sql); do psql languasia languasia -f $i; done;
