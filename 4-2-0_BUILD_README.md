## Building against iRODS 4.2.0

1. Install iRODS 4.2.0

2. Switch to feature/4.2.0_Build branch in GoRODS 
  ```
  $ git checkout "feature/4.2.0_Build"
  ```

3. Follow directions at https://packages.irods.org/ to install new repositories to your applicable distro

4. Install all external packages (irods-externals*) and irods-runtime, irods-devel. 
  ```
  $ sudo yum install irods-externals* irods-runtime irods-devel
  ```

5. Create/edit ~/.irods/irods_environment.json and specify "irods_plugins_home":
  ```
  {
      "irods_plugins_home": "/usr/lib/irods/plugins"
  }
  ```

6. Build your application!

Note: The icommand (or icat/resource) package runtime dependency from 4.1.10 is now found in the irods-runtime package in 4.2.0. Additionally, you must install the irods-externals* packages as runtime dependencies in 4.2.0

| iRODS Version | Build Dependencies | Runtime Dependencies |
| --- | --- | --- |
| <= 4.1.10 | irods-dev | irods-icommands (or irods-icat or irods-resource) |
| >= 4.2.0 | irods-devel, irods-externals* | irods-runtime, irods-externals* | 
