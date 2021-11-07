#! /bin/sh 

### this is why we are in alpine
if ! [ -z "$MM_LICENSE_KEY" ];then
    curl -L "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=$MM_LICENSE_KEY&suffix=tar.gz" -o GeoLite2-City.tar.gz; 
    tar -xzf GeoLite2-City.tar.gz; 
    cp */GeoLite2-City.mmdb ./
    rm -rf GeoLite2-City_* GeoLite2-City.tar.gz
fi

exec "$@"