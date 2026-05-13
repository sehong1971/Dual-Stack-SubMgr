# Subscription Manager

## Subscription Manager User guide

  Subscription manager User guide is available here: https://docs.o-ran-sc.org/projects/o-ran-sc-ric-plt-submgr/en/latest/user-guide.html
  The document source is under the submgr/doc directory. Web document is generated automatically from that after the change has been merged
  in the ORAN Gerrit repo.

## ORAN E2 interface specifications 

  WG3: The Near-Real-Time RIC and E2 Interface Workgroup
  https://www.o-ran.org/specifications

## Compiling Subscription Manager code

### Just to compile Docker image give this command

  docker build -t submgr:tag .

  The command will build the docker image with name 'submgr:tag'. It will run also all the unit tests before image is created.
  Unit tests may fail randomly. There are some cases which may fail due to incorrect timing. in test case. If error happens just
  issue the same command to recompile. Tag value can be anything or it can be omitted.

### Compiling code to run unit tests give this command

  docker build -t submgr . > build.txt

  The command will build the docker image with name 'submgr:latest'. It will run also all the unit tests before image is created.
  Compilation and test cases logs are stores in build.txt. Unit tests may fail randomly. There are some cases which may fail due
  to incorrect timing in test case. If error happens just issue the same command to recompile.

## Subscription Manager ORAN Helm chart

  Subscription Manager ORAN Helm chart is stored here: https://gerrit.o-ran-sc.org/r/admin/repos/ric-plt/ric-dep.

## How to generate E2AP ASN.1 C-codes

  E2AP ASN.1 C-codes has been generated using fork of ASN1C tool. The tool is stored here: https://github.com/nokia/asn1c.

  Latest version of the tool should be used. Follow Build and Install instructions for the tool on the page.

  Generated ASN1 C-codes are stored in submgr/3rdparty/E2AP-v02.00.00 directory. Directory is name based in the specification version.

  Note that any E2AP specifications are not allowed to store in ORAN Gerrit repo! Only generated C-code from the specification can
  be stored there.

  Example: Update current E2AP-v02.00.00 version to E2AP-v02.00.01 version.
  
  Create following directory structure:

    E2AP-v02.00.01
     |
     - spec
        |
        - e2ap-v02.00.01.asn
    
  e2ap-v02.00.01.asn file contains the E2AP v02.00.01 specification's data structures in ASN1 format.

  To generate the ASN.1 C-codes give following command in E2AP-v02.00.01 dir:
  asn1c -pdu=auto -fincludes-quoted -fcompound-names -fno-include-deps -gen-PER -no-gen-OER -no-gen-example spec/e2ap-v02.00.01.asn

  Copy E2AP-v02.00.01 directory with the generated C-codes under submgr/3rdparty directory.

  Remove spec directory under E2AP-v02.00.01 directory.

  Remove old submgr/3rdparty/E2AP-v02.00.00 directory.

## License

  This project is licensed under the Apache License, Version 2.0 - see the LICENSES.txt
