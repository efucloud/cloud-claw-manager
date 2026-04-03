import { MarketApplicationVersionDetail } from "@/services/market_application_version";

export const resetLocalApplication = (
  org: string,
  workspace: string,
  action: string,
  application: MarketApplicationVersionDetail
) => {
  if (action === "create") {
    sessionStorage.removeItem(`application|${org}|${workspace}`);
  } else if (action === "update") {
    sessionStorage.removeItem(
      `application|${org}|${workspace}|${application.id}`
    );
  }
};
export const saveLocalApplication = (
  org: string,
  workspace: string,
  action: string,
  application: MarketApplicationVersionDetail
) => {
  const appString = JSON.stringify(application);
  if (action === "create") {
    sessionStorage.setItem(`application|${org}|${workspace}`, `${appString}`);
  } else if (action === "update") {
    sessionStorage.setItem(
      `application|${org}|${workspace}|${application.id}`,
      `${appString}`
    );
  }
};
export const getLocalApplication = (
  org: string,
  workspace: string,
  action: string,
  id: string
): MarketApplicationVersionDetail => {
  if (action === "create") {
    return JSON.parse(
      sessionStorage.getItem(`application|${org}|${workspace}`) || "{}"
    ) as MarketApplicationVersionDetail;
  } else if (action === "update") {
    return JSON.parse(
      sessionStorage.getItem(`application|${org}|${workspace}|${id}`) || "{}"
    ) as MarketApplicationVersionDetail;
  } else {
    return {} as MarketApplicationVersionDetail;
  }
};
