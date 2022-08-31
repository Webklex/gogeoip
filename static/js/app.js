// Main Application

const api = {
    host: "",

    fetch: (url, options) => {
        return fetch(api.host + url, options).then((response) => response.json())
    },
    statistic: () => {
        return api.fetch("/api/statistic")
    },
    me: () => {
        return api.fetch("/api/me")
    },
    search: (data) => {
        return api.fetch("/api/search?limit=20", {
            method: 'POST',
            headers: {
                'accept': 'application/json',
            },
            body: JSON.stringify(data)
        })
    }
}

const app = {
    dom: {
        statistic: {
            ips: document.getElementById("statistic-ips"),
            asns: document.getElementById("statistic-asns"),
            isps: document.getElementById("statistic-isps"),
            cities: document.getElementById("statistic-cities"),
        },
        loading: document.getElementById("loading"),
        data: document.getElementById("data_holder"),
        map_holder: document.getElementById("map_holder"),
        map: document.getElementById("map"),
        form: document.getElementById("search_form"),
        short_form: document.getElementById("short_form"),
        search_ip: document.getElementById("search_ip"),
        ip: {
            ip: document.getElementById("input_ip_ip"),
            proxy_type: document.getElementById("input_proxy_type"),
            type: document.getElementById("input_type"),
            score: document.getElementById("input_score"),
            threat: document.getElementById("input_threat"),
            user_count: document.getElementById("input_user_count"),
            last_seen: document.getElementById("input_last_seen"),
            is_anonymous: document.getElementById("input_is_anonymous"),
            is_anonymous_proxy: document.getElementById("input_is_anonymous_proxy"),
            is_anonymous_vpn: document.getElementById("input_is_anonymous_vpn"),
            is_hosting_provider: document.getElementById("input_is_hosting_provider"),
            is_public_proxy: document.getElementById("input_is_public_proxy"),
            is_satellite_provider: document.getElementById("input_is_satellite_provider"),
            is_tor_exit_node: document.getElementById("input_is_tor_exit_node"),
            latitude: document.getElementById("input_latitude"),
            longitude: document.getElementById("input_longitude"),
            accuracy_radius: document.getElementById("input_accuracy_radius"),
        },
        country: {
            code: document.getElementById("input_country_code"),
            name: document.getElementById("input_country_name"),
            european_member: document.getElementById("input_country_european_member"),
        },
        continent: {
            code: document.getElementById("input_continent_code"),
            name: document.getElementById("input_continent_name"),
        },
        city: {
            name: document.getElementById("input_city_name"),
            time_zone: document.getElementById("input_city_time_zone"),
            metro_code: document.getElementById("input_city_metro_code"),
            population_density: document.getElementById("input_city_population_density"),
        },
        region: {
            code: document.getElementById("input_region_code"),
            name: document.getElementById("input_region_name"),
        },
        postal: {
            zip: document.getElementById("input_postal_zip"),
        },
        isp: {
            name: document.getElementById("input_isp_name"),
        },
        network: {
            network: document.getElementById("input_network_network"),
            domain: document.getElementById("input_network_domain"),
        },
        organization: {
            name: document.getElementById("input_organization_name"),
        },
        domain: {
            name: document.getElementById("input_domain_name"),
        },
        autonomous_system: {
            name: document.getElementById("input_autonomous_system_name"),
            number: document.getElementById("input_autonomous_system_number"),
        },
    },
    state: {
        show_form: false,
        map_loaded: false,
        map: null,
        data: [],
    },
    nice_number: (number) => {
        number = parseFloat(number);

        if (number < 10) {
            return `${number}`
        }else if (number < 100) {
            return `${number}`
        }else if (number < 1000) {
            return `${number}`
        }else if (number < 10000) {
            return `${(number/1000).toFixed(2)}k`
        }else if (number < 100000) {
            return `${(number/1000).toFixed(1)}k`
        }else if (number < 1000000) {
            return `${(number/1000).toFixed(0)}k`
        }else {
            return `${(number/1000000).toFixed(2)}m`
        }
    },
    fetch_statistic: () => {
        api.statistic().then(res => {
            app.dom.statistic.ips.getElementsByClassName("value")[0].innerHTML = app.nice_number(res.ips);
            app.dom.statistic.asns.getElementsByClassName("value")[0].innerHTML = app.nice_number(res.asns);
            app.dom.statistic.isps.getElementsByClassName("value")[0].innerHTML = app.nice_number(res.isps);
            app.dom.statistic.cities.getElementsByClassName("value")[0].innerHTML = app.nice_number(res.cities);
        }).catch(e => {
        })
    },
    start: () => {
        app.fetch_statistic();

        api.me().then(res => {
            app.state.data = [res.ip];

            app.dom.search_ip.value = res.ip.ip;
            app.dom.ip.ip.value = res.ip.ip;
            app.dom.ip.proxy_type.value = res.ip.proxy_type;
            app.dom.ip.type.value = res.ip.type;
            app.dom.ip.score.value = res.ip.score;
            app.dom.ip.threat.value = res.ip.threat;
            app.dom.ip.user_count.value = res.ip.user_count;
            app.dom.ip.last_seen.value = res.ip.last_seen;
            app.dom.ip.is_anonymous.checked = res.ip.is_anonymous;
            app.dom.ip.is_anonymous_proxy.checked = res.ip.is_anonymous_proxy;
            app.dom.ip.is_anonymous_vpn.checked = res.ip.is_anonymous_vpn;
            app.dom.ip.is_hosting_provider.checked = res.ip.is_hosting_provider;
            app.dom.ip.is_public_proxy.checked = res.ip.is_public_proxy;
            app.dom.ip.is_satellite_provider.checked = res.ip.is_satellite_provider;
            app.dom.ip.is_tor_exit_node.checked = res.ip.is_tor_exit_node;

            app.dom.country.code.value = res.ip.country.code;
            app.dom.country.name.value = res.ip.country.name;
            app.dom.country.european_member.checked = res.ip.country.european_member;

            app.dom.continent.code.value = res.ip.country.continent.code;
            app.dom.continent.name.value = res.ip.country.continent.name;

            app.dom.city.name.value = res.ip.city.name;
            app.dom.city.time_zone.value = res.ip.city.time_zone;
            app.dom.city.metro_code.value = res.ip.city.metro_code;
            app.dom.city.population_density.value = res.ip.city.population_density;

            const region = res.ip.city.regions && res.ip.city.regions.length > 0 ? res.ip.city.regions[0] : {
                code: "",
                name: ""
            };
            app.dom.region.code.value = region.code;
            app.dom.region.name.value = region.name;

            app.dom.isp.name.value = res.ip.isp.name;

            app.dom.network.network.value = res.ip.network.network;
            app.dom.network.domain.value = res.ip.network.domain;

            app.dom.organization.name.value = res.ip.organization.name;

            const domain = res.ip.domains && res.ip.domains.length > 0 ? res.ip.domains[0] : {name: ""};
            app.dom.domain.name.value = domain.name;

            app.dom.autonomous_system.name.value = res.ip.autonomous_system.name;
            app.dom.autonomous_system.number.value = res.ip.autonomous_system.number;

            app.populate_data_holder([res.ip]);
            app.toggle_data_view(true);
        }).catch(e => {
        });

        app.dom.form.onsubmit = function (evt) {
            if (evt) {
                evt.preventDefault();
            }
            app.search()
        }
    },
    search: (e) => {
        app.toggle_data_view(false);

        let data = {};
        if (app.state.show_form) {
            let domains = app.dom.domain.name.value !== "" ? [{
                name: app.dom.domain.name.value,
            }] : [];

            let regions = [];
            if (app.dom.region.code.value !== app.dom.region.name.value) {
                regions = [{
                    code: app.dom.region.code.value,
                    name: app.dom.region.name.value,
                }]
            }

            data = {
                ip: app.dom.ip.ip.value,
                proxy_type: app.dom.ip.proxy_type.value,
                type: app.dom.ip.type.value,
                score: app.dom.ip.score.value,
                threat: app.dom.ip.threat.value,
                user_count: app.dom.ip.user_count.value,
                last_seen: parseInt(app.dom.ip.last_seen.value),
                is_anonymous: app.dom.ip.is_anonymous.checked,
                is_anonymous_proxy: app.dom.ip.is_anonymous_proxy.checked,
                is_anonymous_vpn: app.dom.ip.is_anonymous_vpn.checked,
                is_hosting_provider: app.dom.ip.is_hosting_provider.checked,
                is_public_proxy: app.dom.ip.is_public_proxy.checked,
                is_satellite_provider: app.dom.ip.is_satellite_provider.checked,
                is_tor_exit_node: app.dom.ip.is_tor_exit_node.checked,
                latitude: parseFloat(app.dom.ip.latitude.value),
                longitude: parseFloat(app.dom.ip.longitude.value),
                accuracy_radius: parseFloat(app.dom.ip.accuracy_radius.value),
                domains: domains,
                country: {
                    code: app.dom.country.code.value,
                    name: app.dom.country.name.value,
                    european_member: app.dom.country.european_member.checked,
                    continent: {
                        code: app.dom.continent.code.value,
                        name: app.dom.continent.name.value,
                    },
                },
                city: {
                    name: app.dom.city.name.value,
                    time_zone: app.dom.city.time_zone.value,
                    metro_code: parseInt(app.dom.city.metro_code.value),
                    population_density: parseInt(app.dom.city.population_density.value),
                    regions: regions,
                },
                isp: {
                    name: app.dom.isp.name.value,
                },
                network: {
                    network: app.dom.network.network.value,
                    domain: app.dom.network.domain.value,
                },
                organization: {
                    name: app.dom.organization.name.value,
                },
                autonomous_system: {
                    name: app.dom.autonomous_system.name.value,
                    number: parseInt(app.dom.autonomous_system.number.value),
                },
            };
        } else {
            data = {
                ip: app.dom.search_ip.value,
            }
        }
        api.search(data).then(res => {
            app.state.data = res.rows;
            app.populate_data_holder(res.rows);
            app.draw_map();

            const delta = res.total_rows - res.rows.length;
            if (delta > 0) {
                const elm = document.createElement("div");
                // bg-slate-800 py-4 px-4 md:rounded-lg mt-4
                elm.classList.add("w-full", "bg-slate-800", "py-2", "px-4", "md:rounded-lg", "md:mt-2");
                elm.innerHTML = `
<div class="border border-solid border-1 border-slate-600 md:border-0 py-2 px-2 flex flex-wrap rounded shadow md:shadow-none">
    <div class="flex flex-wrap w-full px-2 text-xs text-center">
            <div class="w-full text-slate-400">${delta} additional ips are available...</div>
            <div class="w-full">Please use the API to query additional ips.</div>
    </div>
</div>`;
                app.dom.data.appendChild(elm)
            }
            app.toggle_data_view(true);
        }).catch(e => {
        });
    },
    toggle_data_view: (state) => {
        if (state === true) {
            app.dom.loading.classList.remove("block");
            app.dom.loading.classList.add("hidden");

            app.dom.data.classList.remove("hidden");
            app.dom.data.classList.add("block");
        } else {
            app.dom.loading.classList.add("block");
            app.dom.loading.classList.remove("hidden");

            app.dom.data.classList.add("hidden");
            app.dom.data.classList.remove("block");
        }
    },
    reset: () => {
        app.dom.form.reset();
        app.dom.search_ip.value = "";
    },
    toggle_form: () => {
        app.dom.form.classList.toggle('hidden');
        app.dom.short_form.classList.toggle('hidden');
        app.state.show_form = !app.state.show_form;
    },
    draw_map_item: (label, longitude, latitude) => {
        if (longitude === latitude && latitude === 0) {
            longitude = 4.35247;
            latitude = 50.84673;
        }

        const iconFeature = new ol.Feature({
            geometry: new ol.geom.Point(ol.proj.fromLonLat([longitude, latitude])),
            name: 'Position',
            population: 4000,
            rainfall: 500
        });
        iconFeature.setStyle(new ol.style.Style({
            image: new ol.style.Icon(({
                anchor: [0.5, 0.5],
                anchorXUnits: 'fraction',
                anchorYUnits: 'pixels',
                scale: .1,
                src: 'https://upload.wikimedia.org/wikipedia/commons/thumb/8/88/Map_marker.svg/156px-Map_marker.svg.png'
            })),
            text: new ol.style.Text({
                text: label,
                font: '18px Calibri,sans-serif',
                fill: new ol.style.Fill({
                    color: '#e2e8f0',
                }),
                stroke: new ol.style.Stroke({
                    color: '#0ea5e9',
                    width: 8,
                }),
            })
        }));

        app.state.map.addLayer(new ol.layer.Vector({
            source: new ol.source.Vector({
                features: [iconFeature]
            })
        }));

        app.state.map.setView(new ol.View({
            center: ol.proj.fromLonLat([longitude, latitude]),
            //maxZoom: 18,
            zoom: 6
        }));
    },
    draw_map: () => {
        app.dom.map.innerHTML = "";
        app.state.map = new ol.Map({
            controls: ol.control.defaults({attribution: false}).extend([new ol.control.Attribution({
                //collapsible: false
            })]),
            layers: [
                new ol.layer.Tile({
                    source: new ol.source.OSM()
                })
            ],
            target: 'map'
        });
        for (let i = 0; i < app.state.data.length; i++) {
            try {
                app.draw_map_item(app.state.data[i].ip, app.state.data[i].longitude, app.state.data[i].latitude);
            } catch (e) {
            }
        }
        app.state.map.render();
    },
    toggle_map: () => {
        app.dom.map_holder.classList.toggle('hidden');

        if (app.dom.map_holder.classList.contains("hidden") === false) {
            // it was hidden
            app.draw_map();
        }
    },
    populate_data_holder: (rows) => {
        if (rows && rows.length > 0) {
            app.dom.data.innerHTML = "";
            for (let i = 0; i < rows.length; i++) {
                app.dom.data.appendChild(app.create_data_element(rows[i]))
            }
        } else {
            app.dom.data.innerHTML = "<div class='bg-slate-800 w-full text-center md:rounded-lg md:mt-1 p-4'>No Data found</div>";
        }
    },
    create_data_element: (data) => {
        const elm = document.createElement("div");
        // bg-slate-800 py-4 px-4 md:rounded-lg mt-4
        elm.classList.add("w-full", "bg-slate-800", "py-2", "px-4", "md:rounded-lg", "md:mt-2");

        elm.innerHTML = `
<div class="border border-solid border-1 border-slate-600 md:border-0 py-2 px-2 flex flex-wrap rounded shadow md:shadow-none">
    <div class="flex flex-wrap w-full">
        <div class="grow px-2">
            <span class="text-xs text-slate-400">IP Address:</span>
            <br />
            <a href="http://${data.ip.includes(':') ? '[' + data.ip + ']' : data.ip}" class="text-sky-300" target="_blank" rel="noreferrer">
                ${data.ip}
            </a> 
        </div>
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">Network:</span>
            <br />
            ${data.network.network || "-"}
        </div>
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">Proxy Type:</span>
            <br />
            ${data.proxy_type || "-"}
        </div>
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">Usage Type:</span>
            <br />
            ${data.type || "-"}
        </div>
        <div class="w-auto px-2 text-center">
            <span class="text-xs text-slate-400">Users:</span>
            <br />
            ${data.user_count || "0"}
        </div>
        <div class="w-auto px-2 text-center">
            <span class="text-xs text-slate-400">Last seen:</span>
            <br />
            ${data.last_seen}
        </div>
        <div class="w-auto px-2 text-center">
            <span class="text-xs text-slate-400">Score:</span>
            <br />
            ${data.score || "0"}
        </div>
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">Threat:</span>
            <br />
            ${app.render_thread_level(data.threat)}
        </div>
    </div>
    <div class="flex flex-wrap w-full">
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">ASN:</span>
            <br />
            ${data.autonomous_system.number || "-"}
        </div>
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">AS Name:</span>
            <br />
            ${data.autonomous_system.name || "-"}
        </div>
    </div>
    <div class="flex flex-wrap w-full md:w-auto">
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">Continent:</span>
            <br />
            ${data.country.continent.name || "-"}
        </div>
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">Country:</span>
            <br />
            ${data.country.name || "-"}
        </div>
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">City:</span>
            <br />
            ${data.city.name || "-"}
        </div>
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">Zip:</span>
            <br />
            ${data.postal.zip || "-"}
        </div>
    </div>
    <div class="flex flex-wrap w-full md:w-auto">
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">Latitude:</span>
            <br />
            ${data.latitude || "-"}
        </div>
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">Longitude:</span>
            <br />
            ${data.longitude || "-"}
        </div>
        <div class="w-auto px-2">
            <span class="text-xs text-slate-400">Accuracy:</span>
            <br />
            ${data.accuracy_radius || "-"}
        </div>
    </div>
</div>`
        return elm;
    },
    render_thread_level: (threat) => {
        switch (threat) {
            case "low":
                return "<span class='bg-green-600 text-slate-900 px-2 py-1 text-xs'>LOW</span>";
            case "mid":
                return "<span class='bg-amber-600 text-slate-900 px-2 py-1 text-xs'>MID</span>";
            case "high":
                return "<span class='bg-red-600 text-slate-900 px-2 py-1 text-xs'>HIGH</span>";
            default:
                return "<span class='bg-sky-600 text-slate-900 px-2 py-1 text-xs'>UNK</span>";
        }
    }
};

(function () {
    app.start();
})();