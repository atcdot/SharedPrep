import { useEffect, useRef } from "react";

interface TelegramLoginProps {
  clientId: string;
  onAuth: (idToken: string) => void;
}

declare global {
  interface Window {
    Telegram?: {
      Login?: {
        init: (
          options: { client_id: string },
          callback: (data: { id_token?: string; error?: string }) => void,
        ) => void;
        open: (callback?: (data: { id_token?: string; error?: string }) => void) => void;
      };
    };
  }
}

export function TelegramLogin({ clientId, onAuth }: TelegramLoginProps) {
  const onAuthRef = useRef(onAuth);
  onAuthRef.current = onAuth;

  useEffect(() => {
    const script = document.createElement("script");
    script.src = "https://oauth.telegram.org/js/telegram-login.js";
    script.async = true;
    script.onload = () => {
      window.Telegram?.Login?.init({ client_id: clientId }, (data) => {
        if (data.id_token) {
          onAuthRef.current(data.id_token);
        }
      });
    };
    document.head.appendChild(script);

    return () => {
      document.head.removeChild(script);
    };
  }, [clientId]);

  const handleOpen = () => {
    window.Telegram?.Login?.open();
  };

  return (
    <button
      onClick={handleOpen}
      className="inline-flex items-center gap-2 rounded-md bg-[#0088cc] px-4 py-2 text-sm font-medium text-white hover:bg-[#006da3] cursor-pointer"
    >
      <svg viewBox="0 0 24 24" className="h-5 w-5 fill-current">
        <path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.479.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z" />
      </svg>
      Log in with Telegram
    </button>
  );
}
