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
  get: function(url) {return send(url, {
    method: 'GET'
  })},
  post: function(url, data) {return send(url, {
    method: 'POST',
    headers: {
      'Contnt-Type': 'application/json'
    },
    body: data
  })},
  put: function(url, data) {return send(url,  {
    method: 'PUT',
    headers: {
      'Contnt-Type': 'application/json'
    },
    body: data
  })},
  delete: function(url) {return send(url, {
    method: 'DELETE'
  })},
  options: function(url) {return send(url, {
    method: 'OPTIONS'
  })}
};