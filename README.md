# Daemon in Go to setup your DNS A Record in AWS Route 53
---
Simple dameon Go program to automatically set your remote IP adress.

It's an own alternative to others providers such as no-ip or dyndns.
## How to use it?

Just put in your system init scripts.

## How install it?

1. Clone/download repository
```
git clone https://github.com/jmrobles/AWSmyIP
```
2. Get dependencies
```
go get
```
3. Set-up your AWS credentials. A recommended way is to create a file in ```$HOME/.aws/credentials``` with your ```aws_access_key_id``` and ```aws_secret_access_key```. For example:
```
[default]
# The access key for your AWS account
aws_access_key_id=<PUT-YOUR-ACCESS-KEY-ID>
# The secret key for your AWS account
aws_secret_access_key=<PUT-YOUR-SECRET-ACCESS-KEY>
```
4. Set full path and environment variables
You need to specify your credentials and AWS Region. For example: us-west-1
```
./AWSmyIP -zoneID <zoneID> -recordSet <recordSet>
```
TODO: create System Service script
