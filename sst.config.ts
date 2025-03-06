/// <reference path="./.sst/platform/config.d.ts" />

export default $config({
    app(input) {
        return {
            name: "vim-guys",
            removal: input?.stage === "production" ? "retain" : "remove",
            protect: ["production"].includes(input?.stage),
            home: "aws",
        };
    },
    async run() {
        const vpc = new sst.aws.Vpc("Vpc", {})
        const cluster = new sst.aws.Cluster("Cluster", {
            vpc,
        })

        const ap = new sst.aws.Service("AuthProxy", {
            // TODO look at usage logs
            scaling: {
                min: 1,
                max: 16,
            },
            image: {
                context: "./auth-proxy/"
            },
            wait: true,
            cluster,
            loadBalancer: {
                rules: [
                    { listen: "80/http", forward: "8080/http" },
                    //{ listen: "443/https", forward: "8080/http" },
                ]
            }
        })

        const vg = new sst.aws.Service("VimGuys", {
            // TODO set these to make it so i can run ~1000 games
            image: {
                context: "./vim-guys/"
            },
            wait: true,
            cluster,
        })
    },
});
