

async function send(Command, Information) {
    const url = new URL("http://localhost:3000/api");
    url.searchParams.append('type', 'files');
    url.searchParams.append('command', Command);
    for (const key in Information) {
        url.searchParams.append(key, Information[key]);  
    }
    console.log(url);
    try {
        const responce = await fetch(url)
        if (!responce.ok) {
            throw new Error(`HTTPerror${responce.status}`);
        }
        const data = await responce.json();
        return data.message;
    } catch (error) {
        console.error(error);
        throw error
    };
};

export const files = {
    read: function(path) {return send("read",{path: path})},
    check: function(path) {return send("check",{path: path})},
    create: function(path) {return send("create",{path: path})},
    delete: function(path) {return send("delete",{path: path})},
    write: function(path,data) {return send("write",{path: path, data:data})},
    rename: function(path,name) {return send("rename",{path: path, name: name})},
    move: function(source,target) {return send("move",{source: source, target:target})},
    copy: function(source,target) {return send("copy",{source: source, target:target})}
};
