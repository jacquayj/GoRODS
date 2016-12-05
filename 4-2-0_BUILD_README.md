## Building against iRODS 4.2.0

1. Install iRODS 4.2.0

2. Switch to feature/4.2.0_Build branch in GoRODS 
  ```
  $ git checkout "feature/4.2.0_Build"
  ```

3. Follow directions at https://packages.irods.org/ to install new repositories to your applicable distro

4. Install all external packages (irods-externals*) and irods-runtime
  ```
  $ sudo yum install irods-externals* irods-runtime
  ```

5. Build your application!
