#if !defined(FACT_H)
#define FACT_H

// http://www.swig.org/tutorial.html section "SWIG for the truly lazy"
// this ensure same module name for ruby require and python import, but
// each runtime share-lib loader expects slightly different names for 
// respective share library runtimes (see Makefile comments) ...

// this is used by the swig pre-processor, rather than using *.i files:
#ifdef SWIG
%module newmath
%{
#define SWIG_FILE_WITH_INIT
#include "fact.h"
%}
#endif

// check for c vs. c++ compiler

#if defined(__cplusplus)
  extern long fact(long n=42);
//  extern int factarray(long* in, long* out, int sz=3);
//  extern int factmatrix(long** in, long** out, int x_sz=3, int y_sz=2);
#else
  extern long fact(long n);
//  extern int factarray(long* in, long* out, int sz);
//  extern int factmatrix(long** in, long** out, int x_sz, int y_sz);
#endif

#endif // FACT_H
