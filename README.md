# loginchecker
This program checks remote logins on a machine. It sends a notification when a logged in user 
or his remote address is not set as allowed in the configuration file. You can set an interval 
the program checks for logins. It will pause notifications for a detected user within the 
configured gracetime.

## configuration
The programm will use `config.yml` in the execution path, you can also set a custom configuration
file by providing the configuration file via commandline argument. 

If you leave the mail or slack key empty, no notification will be sent in this channel. There will always
be an output to stdout of the programm. When you want to enable mail or slack notification be sure to
set all configuration params described below. The basic configuration params have to be set in any way.

## basic configuration parameters

### `allowed_adresses` []string
Set allowed addresses and hostnames which are allowed

### `allowed_users` []string
Set allowed users which are allowed to connect to this machine

### `recheck_time` int64
Interval in seconds to check for logins

### `grace_time` int64
Interval in seconds to resend notification for already detected users

## mail configuration params within key `mail`

### `from` string
Mail from address in email notification header

### `from_name` string
Mail from name in email notification header

### `subject` string
Mail subject in email notification, current hostname will be appended

### `recipients` []string
Mail recipients who should get the notifications

## slack configuration params within key `slack`

### `webhook_url` string
Slack webhook URL for incoming webhook

### `channel` string
Channel to send the notification to

### `author` string
Displayed author of notification

### `message` string
Slack message subject,  current hostname will be appended

### `username` string
Displayed username of notification

### `icon_emoji` string
Slack emoji icon of notification

