## Building against iRODS 4.2.0

1. Install iRODS 4.2.0

1a. Switch to feature/4.2.0_Build branch in GoRODS

```
$ git checkout "feature/4.2.0_Build"
```

2. Follow directions at https://packages.irods.org/ to install new repositories to your applicable distro

3. Install all external packages (irods-externals*) and irods-runtime

```
$ sudo yum install irods-externals* irods-runtime
```

4. Build your application!