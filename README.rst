pinot-exporter
==============


This project exposes some additional metrics that are not exposed by the included Pinot Promentheus metrics

So far the list includes:

- Disk usage per Table (as this is what i did not find)

This is *not* meant as a replacement for metrics that Pinot provides itself, but rather to augment any missing ones.

It uses the REST API to obtain these metrics, so beware depending on the size of your cluster and frequency of polling you requested.

Also be aware that some metrics are cached and not retrieved when Prometheus scrapes this exporter. (as it needs several API calls to get them anyway)


Documentation
-------------

To come.


Authors
--------

* Konstantinos Metaxas 

License
--------

GPL v2. Check LICENSE file
