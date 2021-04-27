# README

This is a simple application I wrote which cleans up phantom clients in the [UniFi Controller](https://help.ui.com/hc/en-us/sections/360008076434-Using-the-UniFi-Network-Controller). I've been having problems over the past couple years with my UniFi Controller registering "phantom" clients. These clients have random MAC addresses and have no network activity. They just appear. I'm not the only one with this problem:

- [https://www.reddit.com/r/Ubiquiti/comments/edpia4/unknown_devices_appearing_in_client_list](https://www.reddit.com/r/Ubiquiti/comments/edpia4/unknown_devices_appearing_in_client_list/)
- [https://www.reddit.com/r/Ubiquiti/comments/9bdu8s/unknown_mac_addresses_in_unifi/](https://www.reddit.com/r/Ubiquiti/comments/9bdu8s/unknown_mac_addresses_in_unifi/)

I run Docker containers within my home network for various functions, including [Home Assistant](https://github.com/rfpludwick/home-assistant-config). I believe these phantom clients are a result of that setup. And a byproduct of these phantom clients is that, since I'm running Home Assistant and have integrated it with my UniFi Controller, these clients make their way into my Home Assistant entities. Manual cleanup in **both** systems is extremely tedious and annoying.

Thus, this project. It's my first project in Go, so please, don't judge me too harshly; I know some of the approach here is pretty simplistic. All it does is logs in to the UniFi Controller, gets the list of clients, and checks them for any network activity or custom name. If they have no network activity and no custom name, then they are summarily removed from the UniFi Controller.

If you want to use this, by all means, but I take no responsibility if it causes problems with your UniFi Controller. Just download the source, build it for your target, and run it. It'll need a configuration JSON (defaults to reading `configuration.json` in the same directory as the executable):

```json
{
    "host": "",
    "site": "",
    "username": "",
    "password": ""
}
```

And invoke it so (Linux example here):

```shell
./<executable>
```

That's it. It implements *getopt* as well, so if you want to see what options are available:

```shell
./<executable> -h
```
