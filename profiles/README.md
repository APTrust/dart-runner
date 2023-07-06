# DART BagIt Profiles

Except for `btr-v1.0-1.3.0.json` profiles in this directory are built into DART and Dart Runner, and are used in testing.

* `aptrust-v2.2.json` is the official APTrust profile
* `btr-v1.0.json` was the official Beyond the Repository (BTR) profile included in DART since 2018 or so. It's actually a pre-release of the BTR profile. See below for more info.
* `empty_profile.json` is an empty profile that simply conforms to the minimum BagIt spec. This profile has two purposes. 1) It can be used in validation to tell you whether any bag conforms to the minimum BagIt spec. 2) You can use it as a starter template to build your own profiles in the DART UI.

The template `btr-v1.0-1.3.0.json` became the official BTR spec after the BagIt Profile standard moved from version 1.2.0 to version 1.3.0. This profile points to a different URL than `btr-v1.0.json` and includes a "recommended" attribute on some tags. It also includes some additional fields in the BagItProfileInfo section. 
