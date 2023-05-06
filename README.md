# goetotp

`goetotp` is [et-otp](http://ecki.github.io/et-otp/) alternative command line tool that generate TOTP code.

## Usage

* Set up with `et-otp.jar` and create `.et-top.properties` file in your directory

```sh
$  goetotp --unlockpassword <YOUR UNLOCK PASSWORD in et-otp.jar>
```

### Example

```sh
$  goetotp --unlockpassword MyPassword9999
123456
```

You can use password via environment variable or std input.

```sh
$ goetotp
Enter unlock password: 
123456

$ export ETOTP_PASSWORD=xxxxxxxxxxxxxxxx
$ goetotp
123456
```

### Installation

From source code:

```sh
go install github.com/ma91n/goetotp/cmd/goetotp@latest
```
