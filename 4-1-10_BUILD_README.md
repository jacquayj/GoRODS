## Installation - iRODS 4.1.10

These instructions are for building against iRODS 4.1.10 packages. See the [4.2.0 build considerations README](https://github.com/jjacquay712/GoRODS/blob/master/4-2-0_BUILD_README.md) for information on building against 4.2.0.

**Step #1:**  Switch to feature/4.1.10_Build branch in GoRODS 
  ```
  $ git checkout "feature/4.1.10_Build"
  ```

**Step #2:** Install build dependencies (http://irods.org/download/): irods-dev-4.1.10

```
CentOS/RHEL (64 bit)
$ sudo yum install ftp://ftp.renci.org/pub/irods/releases/4.1.10/centos7/irods-dev-4.1.10-centos7-x86_64.rpm

Ubuntu (64 bit)
$ curl ftp://ftp.renci.org/pub/irods/releases/4.1.10/ubuntu14/irods-dev-4.1.10-ubuntu14-x86_64.deb > irods-dev-4.1.10-ubuntu14-x86_64.deb
$ sudo dpkg -i irods-dev-4.1.10-ubuntu14-x86_64.deb
```

**Step #3:** Install runtime dependencies (http://irods.org/download/): irods-icommands-4.1.10

```
CentOS/RHEL (64 bit)
$ sudo yum install ftp://ftp.renci.org/pub/irods/releases/4.1.10/centos7/irods-icommands-4.1.10-centos7-x86_64.rpm

Ubuntu (64 bit)
$ curl ftp://ftp.renci.org/pub/irods/releases/4.1.10/ubuntu14/irods-icommands-4.1.10-ubuntu14-x86_64.deb > irods-icommands-4.1.10-ubuntu14-x86_64.deb
$ sudo dpkg -i irods-icommands-4.1.10-ubuntu14-x86_64.deb
```

**Note:** The irods-icat-4.1.10 or irods-resource-4.1.10 packages also contain the required /var/lib/irods/plugins/network/libtcp.so shared object that is loaded at runtime. Be sure that at least one of those three packages is installed when deploying a GoRODS binary.