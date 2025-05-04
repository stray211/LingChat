async function send(url, information) {
  try {
    const responce = await fetch(url, information)
    if (!responce.ok) {
      throw new Error(responce.status);
    }
    return await responce.json();
  } catch (error) {
    console.error(error);
    throw error
  };
};

export const api = {
  get: function(url,data) {
    url = new URL(url);
    for (const key in data) {
      url.searchParams.append(key, data[key]);  
    };
    return send(url, {
      method: 'GET'
  })},
  post: function(url, data) {
    return send(url, {
      method: 'POST',
      headers: {'Contnt-Type': 'application/json'},
      body: data
  })}
};