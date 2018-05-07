## Changelog
### [v0.4.0](https://github.com/dansteen/consuldog/releases/tag/v0.4.0) - 07 May 2018 
Change things so that consuldog now pulls monitor templates from a provided url rather than requiring the files to be present on the host in advance.

Also use spaces as delimiter for consul service tag rather than colons (to allow ports)

#### Breaking Changes
The second field in the tag set on a service should now be a URI to a template file.  This will be pulled in by consuldog itself and parsed as a template.   As part of this the 'templateFile' cli flag is no longer supported.  However file:// URIs are supported so the original behavior is still supported by specifying file://<path_to_template>

Also use spaces as delimiter for consul service tag rather than colons (to allow ports)

#### Bug Fixes

#### Improvements


### [v0.3.0](https://github.com/dansteen/consuldog/releases/tag/v0.3.0) - 02 Oct 2017 
Allow each service to have more than one monitor attached to it

#### Breaking Changes

#### Bug Fixes

#### Improvements

### [v0.2.6](https://github.com/dansteen/consuldog/releases/tag/v0.2.6) - 01 Aug 2017 
Fixed issues with datadog process detection and signaling

#### Breaking Changes

#### Bug Fixes

#### Improvements

### [v0.2.4](https://github.com/dansteen/consuldog/releases/tag/v0.2.4) - 31 Jul 2017
Fixed issues with yaml template processing

#### Breaking Changes

#### Bug Fixes

#### Improvements

### [v0.2.1](https://github.com/dansteen/consuldog/releases/tag/v0.2.1) - 28 Jul 2017
Initial release

#### Breaking Changes

#### Bug Fixes

#### Improvements


