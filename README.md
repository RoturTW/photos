# photos

This is a photo hosting service built using OSL.go.

## Config

```json
{
    "quotas": {
        "mist": 10737418240
    },
    "useSubscriptions": true,
    "subscriptionSizes": {
        "drive": 104857600,
        "pro": 1073741824,
        "max": 10737418240
    }
}
```

- `quotas`: A map of usernames to quotas in bytes. This is used to determine whether a user can upload or not.
- `useSubscriptions`: Whether to use subscriptions or not. If this is set to `false`, it will ignore rotur subscription quotas.
