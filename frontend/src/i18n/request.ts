import { getRequestConfig } from 'next-intl/server';

export default getRequestConfig(async () => {
  const locale = 'pt-BR';

  return {
    locale,
    timeZone: 'America/Sao_Paulo',
    messages: (await import(`../../messages/${locale}.json`)).default
  };
});
