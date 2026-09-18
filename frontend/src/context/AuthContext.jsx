import { createContext, useCallback, useEffect, useState } from "react";
import { authApi } from "../services/api";

export const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    authApi
      .me()
      .then((data) => setUser(data))
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  const signup = useCallback(async (email, password, confirmPassword) => {
    const data = await authApi.signup(email, password, confirmPassword);
    if (data?.token) {
      localStorage.setItem("livepoll_auth_token", data.token);
    }
    setUser(data);
    return data;
  }, []);

  const login = useCallback(async (email, password) => {
    const data = await authApi.login(email, password);
    if (data?.token) {
      localStorage.setItem("livepoll_auth_token", data.token);
    }
    setUser(data);
    return data;
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } catch {
      // ignore
    }
    localStorage.removeItem("livepoll_auth_token");
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, signup, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
