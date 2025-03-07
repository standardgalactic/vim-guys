/// <reference path="./.sst/platform/config.d.ts" />

import { profile } from "console";

export default $config({
    app(input) {
        return {
            name: "vim-guys",
            removal: input?.stage === "production" ? "retain" : "remove",
            protect: ["production"].includes(input?.stage),
            home: "aws",
            providers: { cloudflare: "5.49.1" },
        };
    },
    async run() {
        const vpc = new sst.aws.Vpc("Vpc", {});
        const cluster = new sst.aws.Cluster("Cluster", {
            vpc,
        });

        const prod = $app.stage === "production"
        const domain = prod
            ? "vimguys.theprimeagen.com"
            : `${$app.stage}.vimguys.theprimeagen.com`;

        const ap = new sst.aws.Service("AuthProxy", {
            // TODO look at usage logs
            scaling: {
                min: 1,
                max: 16,
            },
            image: {
                context: "./auth-proxy/",
            },
            wait: true,
            cluster,
            loadBalancer: {
                domain: {
                    name: domain,
                    dns: sst.cloudflare.dns({
                        proxy: prod,
                    }),
                },
                rules: [
                    { listen: "80/http", forward: "8080/http" },
                    { listen: "443/https", forward: "8080/http" },
                ],
            }
        });
        const vg = new sst.aws.Service("VimGuys", {
            // TODO set these to make it so i can run ~1000 games
            image: {
                context: "./vim-guys/",
            },
            wait: true,
            cluster,
        });
    },
});

