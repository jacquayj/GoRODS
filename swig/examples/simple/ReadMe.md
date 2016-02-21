This demonstrates how one can make existing C code available
to Ruby or Python (and many other runtimes) as an external module.

The Makefile comments provide some info. on swig's idiosynchracies.
To learn more about swig, please see http://swig.org. For Ruby and
Python specifics, take a look at http://www.swig.org/Doc3.0/Ruby.html
and http://www.swig.org/Doc3.0/Python.html.

Use "make" to build target modules; and "make test" to verify.
Edit the make macros in the Makefile to point to the desired
installation(s) of Ruby and Python. The initial version of the
Makefile sets the macros to "/opt/include" and "/opt/lib" etc.

Tested on Ubuntu. Needs testing on other Linuxes and MAC OS.
