function fn() {
  const envVars = {};
  // Get all environment variables from the OS
  const System = Java.type('java.lang.System');
  const env = System.getenv();
  const keys = env.keySet().toArray();
  for (let i = 0; i < keys.length; i++) {
      const key = keys[i];
      envVars[key] = env.get(key);
  }
  // Add Karate's own env variable
  envVars['karate.env'] = karate.env;
  const config = {
      karate: {
          properties: {...envVars
          //,additionalProperty: 'value'
          },
      }
  };
  return config;
}