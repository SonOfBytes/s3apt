# apt-method-s3 or s3apt

### Table of Contents
1. [License](#license)
1. [Requirements](#requirements)
1. [Configuration](#configuration)
1. [Usage](#usage)
1. [Testing](#testing)
1. [Contribution](#contribution)

## apt-method-s3
Allow to have a privately hosted apt repository on S3. Access keys are read from
any standard AWS credentials source.  If an instance role is used then the s3 method
transport will attempt to determine the AWS regions auto-magically from the metsadata.

When using an instance role for credentials to the s3 bucket no other configuration is
required.

## License

See LICENSE file.

## Requirements
### Additional package dependencies (except installed by default in Debian)
1. ca-certificates is required

## Configuration

None

## Usage
Install the .deb package from the releases page.  The bucket repo should be
specified using an s3:// prefix, for example:

`deb s3://aptbucketname/repo/ trusty main contrib non-free`

## Testing
The module will run in interactive mode.  It accepts on `stdin` and outputs on `stdout`.  The messages it accepts on stdin
are in the following format and [documented here](http://www.fifi.org/doc/libapt-pkg-doc/method.html/index.html#abstract).

```
600 URI Acquire
URI:s3://my-s3-repository/project-a/dists/trusty/main/binary-amd64/Packages
Filename:Packages.downloaded
Fail-Ignore:true
Index-File:true
```

This message will trigger an s3 get from the above bucket and key and save it to Filename.

## Contribution
If you want to contribute a PR please do so - if the work is substantial I advise talking about it first as an opened issue.
