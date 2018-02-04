#!/bin/bash


#----- Options
dopackage=true
verbose=false


#----- Process arguments
for i in "$@"
do
   case ${i} in
      "-skipPackage")
         echo "Skipping package creation"
         dopackage=false
         shift
         ;;
      "-v"|"-verbose")
         echo "Enabling verbose output"
         verbose=true
         shift
         ;;         
      *)
         echo "Invalid option: ${i}"
         ;;
   esac
done


#----- MacOS
echo
tput bold
echo "Build target: MacOS (64-bits)"
tput sgr0
# Clean up old targets, if any
echo "- cleaning up old targets"
rm -rf macos 2> /dev/null
rm nettest-macos.tar.gz 2> /dev/null
# Compile for target platform
echo "- compiling and building"
if $verbose ; then
    GOOS=darwin GOARCH=amd64 go build -a -v -o macos/nettest *.go
else
    GOOS=darwin GOARCH=amd64 go build -o macos/nettest *.go
fi
# Create tarball for platform
if $dopackage ; then
    echo "- creating archive"
    cp resources/config.yaml macos/
    tar czf nettest-macos.tar.gz macos/* > /dev/null
    echo
fi


#----- Windows
echo
tput bold
echo "Build target: Windows (64-bits)"
tput sgr0
# Clean up old targets, if any
echo "- cleaning up old targets"
rm -rf windows 2> /dev/null
rm nettest-windows.zip 2> /dev/null
# Compile for target platform
echo "- compiling and building"
if $verbose ; then
    GOOS=windows GOARCH=amd64 go build -a -v -o windows/nettest.exe *.go
else
    GOOS=windows GOARCH=amd64 go build -o windows/nettest.exe *.go
fi
# Create tarball for platform
if $dopackage ; then
    echo "- creating archive"
    cp resources/config.yaml windows/
    zip nettest-windows.zip windows/* > /dev/null
    echo
fi


#----- Linux
echo
tput bold
echo "Build target: Linux (64-bits)"
tput sgr0
# Clean up old targets, if any
echo "- cleaning up old targets"
rm -rf linux 2> /dev/null
rm nettest-linux.tar.gz 2> /dev/null
# Compile for target platform
echo "- compiling and building"
if $verbose ; then
    GOOS=linux GOARCH=amd64 go build -a -v -o linux/nettest *.go
else
    GOOS=linux GOARCH=amd64 go build -o linux/nettest *.go
fi
# Create tarball for platform
if $dopackage ; then
    echo "- creating archive"
    cp resources/config.yaml linux/
    tar czf nettest-linux.tar.gz linux/* > /dev/null
    echo
fi


#----- Wrap up
echo
echo "All done."
if $dopackage ; then
    echo "Created archives:"
    ls -1 *.tar.gz *.zip
fi
echo
